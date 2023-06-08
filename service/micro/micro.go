package micro

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/base/event"
	r "github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/db/mongo"
	ed "github.com/goonma/sdk/eventdriven"
	"github.com/goonma/sdk/jwt"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type Micro struct {
	Id string
	//key-value store management
	Config *vault.Vault
	//db map[string]dbconnection
	Mgo db.MongoDB
	//redis
	Rd r.CacheHelper
	//publisher event
	Pub map[string]*ed.EventDriven
	//micro client
	Client map[string]*MicroClient
	//
	two_FA_Key string
	token_Key  string

	//detaul init redis cache
	init_redis bool
}

/*
args[0]: model list
args[1]: not exist || exist && true then initial publisher, else don't implement publisher
args[2]: micro client list map[string]string (name - endpoint address)
*/

func (micro *Micro) Initial(config *vault.Vault, args ...interface{}) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn("Can not get Hostname :"+err.Error(), "MICRO", "HOST_NAME")
		micro.Id = config.GetServiceName()
	} else {
		micro.Id = fmt.Sprintf("%s-%s", config.GetServiceName(), hostname)
	}
	//config store
	micro.Config = config
	//initial DB args[0] => mongodb
	if len(args) > 0 {
		if args[0] != nil {
			models, err := utils.ItoDictionary(args[0])
			if err != nil {
				log.ErrorF(err.Error(), "MICRO", "INITIAL_CONVERTION_MODEL")
			} else {
				err_init := micro.Mgo.Initial(micro.Config, models)
				if err_init != nil {
					log.Warn(err_init.Msg(), err_init.Group(), err_init.Key())
				}
			}
		}
	}
	//read 2FA Key for verify token
	micro.two_FA_Key = micro.Config.ReadVAR("key/2fa/KEY")
	micro.token_Key = micro.Config.ReadVAR("key/api/KEY")
	//initial Event
	service_path := strings.ReplaceAll(micro.Config.GetServiceName(), ".", "/")
	if len(args) > 1 {
		c, err := utils.ItoBool(args[1])
		if err != nil {
			log.Warn("Convert Iterface to Bool :"+err.Error(), "MICRO", "HOST_NAME")
		}
		if utils.Type(args[1]) == "bool" && c {
			//find publisher list
			check, err_p := micro.Config.CheckPathExist(service_path + "/pub/kafka")
			if err_p != nil {
				log.ErrorF(err_p.Msg(), micro.Config.GetServiceName(), "Initial")
			}
			micro.Pub = make(map[string]*ed.EventDriven)
			if check { //custom publisher, list event
				event_list := micro.Config.ListItemByPath(service_path + "/pub/kafka")
				for _, event := range event_list {
					if !Map_PublisherContains(micro.Pub, event) && event != "general" {
						micro.Pub[event] = &ed.EventDriven{}
						//micro.Pub[event].SetNoUpdatePublishTime(true)
						err := micro.Pub[event].InitialPublisherWithGlobal(micro.Config, fmt.Sprintf("%s/%s/%s", service_path, "pub/kafka", event), micro.Config.GetServiceName(), event)
						if err != nil {
							log.ErrorF(err.Msg(), micro.Config.GetServiceName(), "Initial")
						}
					}
				}
			} else { //
				path := "eventbus/kafka"
				check, err := micro.Config.CheckItemExist(service_path + "/pub/kafka")
				if err == nil && check { //if pub/kafka is object not folder then use it, else use main bus
					path = service_path + "/pub/kafka"
				}
				micro.Pub["main"] = &ed.EventDriven{}
				err_p := micro.Pub["main"].InitialPublisher(micro.Config, path, micro.Id)
				if err_p != nil {
					log.ErrorF(err_p.Msg(), err_p.Group(), err_p.Key())
				}
			}
		}
	}
	//len(args)==3: models, kafka, Client
	if len(args) > 2 {
		if args[2] != nil {
			remote_services, err := utils.ItoDictionaryS(args[2])
			if err != nil {
				log.ErrorF(err.Error(), "MICRO", "INITIAL_CONVERTION_MODEL")
			} else {
				micro.InitialMicroClient(remote_services, false)
			}
		}
	}

	//init redis
	if micro.init_redis {
		fmt.Println("===Init Redis===")
		redis, err_r := r.NewCacheHelper(micro.Config,service_path+"/redis")
		if err_r != nil {
			log.ErrorF(err_r.Msg())
		}
		micro.Rd = redis
	}
}

func (micro *Micro) InitialMicroClient(remote_services map[string]string, initConnection bool) {
	//
	micro.Client = make(map[string]*MicroClient)
	//
	var err *e.Error
	for k, v := range remote_services {
		if k != "" {
			if initConnection {
				micro.Client[k], err = NewMicroClient(v)
				if err != nil {
					log.ErrorF(err.Msg(), err.Group(), err.Key())
				} else {
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s", k, v, " Initial success"))
				}
			} else {
				micro.Client[k], err = NewMicroClientWithoutConnection(v)
				if err != nil {
					log.ErrorF(err.Msg(), err.Group(), err.Key())
				} else {
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s", k, v, " Initial success"))
				}
			}
		}
	}
	//
}

func (micro *Micro) SetInitRedis(i bool) {
	micro.init_redis = i
}

func (micro *Micro) PushEvent(ev event.Event) *e.Error {
	event := event.Event{
		EventID:      uuid.New(),
		EventName:    ev.EventName,
		EventData:    ev.EventData,
		Uid:          ev.Uid,
		SourceID:     micro.Id,
		Flow:         ev.Flow,
		PushlishTime: time.Now(),
	}
	if Map_PublisherContains(micro.Pub, ev.EventName) && ev.EventName != "general" {
		return micro.Pub[ev.EventName].Publish(event)
	} else {
		return micro.Pub["main"].Publish(event)
	}

}

func (micro *Micro) GetID() string {
	return micro.Id
}

func (micro *Micro) GetDB(col_name string) (*mongo.Collection, *e.Error) {
	if micro.Mgo.Cols[col_name] != nil {
		return micro.Mgo.Cols[col_name], nil
	} else {
		return nil, e.New(fmt.Sprintf(" DB not initial: %s", col_name), "Micro", "GetDB")
	}
}
func (micro *Micro) GetClient(client_name string) (*MicroClient, *e.Error) {
	if micro.Client[client_name] != nil {
		return micro.Client[client_name], nil
	} else {
		return nil, e.New(fmt.Sprintf(" Micro Client not initial: %s", client_name), "Micro", "GetClient")
	}
}

func (micro *Micro) GetUserID(ctx context.Context) string {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return ""
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key, token)
	if err_v != nil {
		return ""
	}
	return claims.UserID
}

func (micro *Micro) GetJWTContent(ctx context.Context) *jwt.CustomClaims {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return nil
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key, token)
	if err_v != nil {
		return nil
	}
	return claims
}

func (micro *Micro) GetRoleID(ctx context.Context) int {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return 0
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key, token)
	if err_v != nil {
		return 0
	}
	return claims.RoleID
}
func (micro *Micro) IsUserVerify(ctx context.Context) int {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return 0
	}
	claims, err_v := jwt.VerifyJWTToken(micro.token_Key, token)
	if err_v != nil {
		return 0
	}
	return claims.IsVerify
}
func (micro *Micro) GetToken(ctx context.Context) string {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return ""
	}
	_, err_v := jwt.VerifyJWTToken(micro.token_Key, token)
	if err_v != nil {
		return ""
	}
	return token
}
func Map_PublisherContains(m map[string]*ed.EventDriven, item string) bool {
	if len(m) == 0 {
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}
