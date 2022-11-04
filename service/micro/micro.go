package micro
import (
	"strings"
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	"github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/db/mongo"
	"github.com/goonma/sdk/utils"
	e "github.com/goonma/sdk/base/error"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/goonma/sdk/jwt"
	"context"
	"github.com/google/uuid"
	"time"
	"os"
	"fmt"
)

type Micro struct{
	Id string
	//key-value store management
	Config *vault.Vault
	//db map[string]dbconnection
	Mgo  db.MongoDB
	//publisher event
	Pub map[string]*ed.EventDriven
	//micro client
	Client map[string]*MicroClient
	//
	two_FA_Key  string
	token_Key  string
}
/*
args[0]: model list
args[1]: not exist || exist && true then initial publisher, else don't implement publisher 
args[2]: micro client list map[string]string (name - endpoint address)
*/

func (micro *Micro) Initial(config *vault.Vault,args...interface{}){
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn("Can not get Hostname :"+err.Error(),"MICRO","HOST_NAME")
		micro.Id=config.GetServiceName()
	}else{
		micro.Id=fmt.Sprintf("%s-%s",config.GetServiceName(),hostname)
	}
	//config store
	micro.Config=config
	//initial DB args[0] => mongodb
	if len(args)>0{
		if args[0]!=nil{
			models,err:=utils.ItoDictionary(args[0])
			if err!=nil{
				log.ErrorF(err.Error(),"MICRO","INITIAL_CONVERTION_MODEL")
			}else{
				err_init:=micro.Mgo.Initial(micro.Config,models)
				if err_init!=nil{
					log.Warn(err_init.Msg(),err_init.Group(),err_init.Key())
				}
			}
		}
	}
	//read 2FA Key for verify token
	micro.two_FA_Key=micro.Config.ReadVAR("key/2fa/KEY")
	micro.token_Key=micro.Config.ReadVAR("key/api/KEY")
	//initial Event
	if len(args)>1{
		c,err:=utils.ItoBool(args[1])
		if err!=nil{
			log.Warn("Convert Iterface to Bool :"+err.Error(),"MICRO","HOST_NAME")
		}
		if utils.Type(args[1])=="bool" && c{
			service_path:=strings.ReplaceAll(micro.Config.GetServiceName(),".","/")
			//find publisher list
			check,err_p:=micro.Config.CheckPathExist(service_path+"/pub/kafka")
			if err_p!=nil{
				log.ErrorF(err_p.Msg(),micro.Config.GetServiceName(),"Initial")
			}
			if check{//custom publisher, list event
				event_list:=micro.Config.ListItemByPath(service_path+"/pub/kafka")
				micro.Pub=make(map[string]*ed.EventDriven)
				for _,event:=range event_list{
					if !Map_PublisherContains(micro.Pub,event) && event!="general"{
						micro.Pub[event]=&ed.EventDriven{}
						//micro.Pub[event].SetNoUpdatePublishTime(true)
						err:=micro.Pub[event].InitialPublisherWithGlobal(micro.Config,fmt.Sprintf("%s/%s/%s",service_path,"/pub/kafka",event),micro.Config.GetServiceName(),event)
						if err!=nil{
							log.ErrorF(err.Msg(),micro.Config.GetServiceName(),"Initial")
						}
					}
				}
			}else{//use main bus
				micro.Pub["main"]=&ed.EventDriven{}
				err_p:=micro.Pub["main"].InitialPublisher(micro.Config,"eventbus/kafka",micro.Id)
				if err_p!=nil{
					log.ErrorF(err_p.Msg(),err_p.Group(),err_p.Key())
				}
			}
		}
	}
	//len(args)==3: models, kafka, Client
	if len(args)>2{
		if args[2]!=nil{
			remote_services,err:=utils.ItoDictionaryS(args[2])
			if err!=nil{
				log.ErrorF(err.Error(),"MICRO","INITIAL_CONVERTION_MODEL")
			}else{
				micro.InitialMicroClient(remote_services,false)
			}
		}
	}
}
func (micro *Micro) InitialMicroClient(remote_services map[string]string,initConnection bool){
	//
	micro.Client=make(map[string]*MicroClient)
	//
	var err *e.Error
	for k,v:=range remote_services{
		if k!=""{
			if initConnection{
				micro.Client[k],err=NewMicroClient(v)
				if err!=nil{
					log.ErrorF(err.Msg(),err.Group(),err.Key())
				}else{
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s",k,v," Initial success"))
				}
			}else{
				micro.Client[k],err=NewMicroClientWithoutConnection(v)
				if err!=nil{
					log.ErrorF(err.Msg(),err.Group(),err.Key())
				}else{
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s",k,v," Initial success"))
				}
			}
		}
	}
	//
}
func (micro *Micro)PushEvent(ev event.Event) *e.Error{
	event:=event.Event{
		EventID    : uuid.New(),
		EventName  : ev.EventName,
		EventData  : ev.EventData,
		Uid		   : ev.Uid,
		SourceID   : micro.Id,
		Flow       : ev.Flow,
		PushlishTime: time.Now(),
	}
	if !Map_PublisherContains(micro.Pub,ev.EventName) && ev.EventName!="general"{
		return micro.Pub[ev.EventName].Publish(event)
	}else{
		return micro.Pub["main"].Publish(event)
	}
	
}

func (micro *Micro)GetID() string{
	return micro.Id
}

func (micro *Micro)GetDB(col_name string) (*mongo.Collection,*e.Error){
	if micro.Mgo.Cols[col_name]!=nil{
		return micro.Mgo.Cols[col_name],nil
	}else{
		return nil,e.New(fmt.Sprintf(" DB not initial: %s",col_name),"Micro","GetDB")
	}
}
func (micro *Micro)GetClient(client_name string) (*MicroClient,*e.Error){
	if micro.Client[client_name]!=nil{
		return micro.Client[client_name],nil
	}else{
		return nil,e.New(fmt.Sprintf(" Micro Client not initial: %s",client_name),"Micro","GetClient")
	}
}

func (micro *Micro)GetUserID(ctx context.Context) string{
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return ""
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key,token)
	if err_v != nil {
		return ""
	}
	return claims.UserID 
}

func (micro *Micro)GetRoleID(ctx context.Context) int{
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return 0
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key,token)
	if err_v != nil {
		return 0
	}
	return claims.RoleID
}
func (micro *Micro)GetToken(ctx context.Context) string{
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return ""
	}
	_, err_v := jwt.VerifyJWTToken(micro.token_Key,token)
	if err_v != nil {
		return ""
	}
	return token
}
func Map_PublisherContains(m map[string]*ed.EventDriven, item string) bool {
	if len(m)==0{
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}
	