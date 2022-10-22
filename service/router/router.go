package router

import (
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	"github.com/goonma/sdk/log"
	ev "github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/cache/redis"
	//"github.com/goonma/sdk/pubsub/kafka"
	//"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/utils"
	e "github.com/goonma/sdk/base/error"
	//"github.com/goonma/sdk/db/mongo/status"
	"github.com/goonma/sdk/pubsub/kafka"
	//e "github.com/goonma/sdk/base/error"
	"os"
	"github.com/joho/godotenv"
	"github.com/goonma/sdk/db"
	"fmt"
	//"errors"
)

type Router struct {
	//router name
	router_name string
	//key-value store management
	config *vault.Vault
	//event driven
	Ed ed.EventDriven
	//
	Mgo  db.MongoDB
	//destination events bus
	Pub map[string]*ed.EventDriven
	Redis redis.CacheHelper
	TTL int
}

//initial n router
func (r *Router) Initial(router_name string,callbackfn ev.ConsumeFn,args...interface{}) {
	//get ENV
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	log.Initial(router_name)
	//initial Server configuration
	var config vault.Vault
	r.config= &config
	r.config.Initial(fmt.Sprintf("%s/%s","worker",router_name))
	//
	//ReInitial Destination for Logger
	if log.LogMode()!=2{// not in local, locall just output log to std
		log_dest:=r.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest=="kafka"{
			config_map:=kafka.GetConfig(r.config,"logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//init TTL
	c,err_ttl:=r.config.CheckItemExist("router/config")
	if err_ttl!=nil{
		log.Warn(err_ttl.Msg(),"Router","Initial")
	}
	if c{	
		ttl_str:=r.config.ReadVAR("router/config/TTL")
		if ttl_str==""{
			r.TTL=3 //default
		}else{
			ttl_num:=utils.StringToInt(r.config.ReadVAR("router/config/TTL"))
			if ttl_num<0{
				r.TTL=3//default value
			}else{
				r.TTL=ttl_num
			}
		}

	}else{
		r.TTL=3 //default TTL
	}
	//log.Info(fmt.Sprintf("Router TTL: %d",r.TTL),"")
	//initial Event subscriber
	err_s:=r.Ed.InitialSubscriber(r.config,"router/sub/kafka","router",callbackfn,nil)
	if err_s!=nil{
		log.ErrorF(err_s.Msg(),"Router","Initial")
	}
	// load event_routing(event_name, bus_name) table/ event_mapping db
	check,err_p:=r.config.CheckPathExist("router/pub/kafka")
	if err_p!=nil{
		log.ErrorF(err_p.Msg(),"Router","Initial")
	}
	if check{
		event_list:=r.config.ListItemByPath("router/pub/kafka")
		r.Pub=make(map[string]*ed.EventDriven)
		for _,event:=range event_list{
			if !Map_PublisherContains(r.Pub,event) && event!="general"{
				r.Pub[event]=&ed.EventDriven{}
				r.Pub[event].SetNoUpdatePublishTime(true)
				err:=r.Pub[event].InitialPublisherWithGlobal(r.config,fmt.Sprintf("%s/%s","router/pub/kafka",event),"Router",event)
				if err!=nil{
					log.ErrorF(err.Msg(),"Router","Initial")
				}
			}
		}
	}
	//initial Event publisher base number of bus_name

	//cache redis check uid before routing
	var errc *e.Error
	r.Redis,errc=redis.NewCacheHelper(r.config)
	if errc!=nil{
		log.ErrorF(errc.Msg(),"Router","Initial")
	}
	//
	//initial DB args[0] => mongodb
	if len(args)>0{
		if args[0]!=nil{
			models,err:=utils.ItoDictionary(args[0])
			if err!=nil{
				log.ErrorF(err.Error(),"MICRO","INITIAL_CONVERTION_MODEL")
			}else{
				err_init:=r.Mgo.Initial(r.config,models)
				if err_init!=nil{
					log.Warn(err_init.Msg(),err_init.Group(),err_init.Key())
				}
			}
		}
	}
}

//start consumers
func (r *Router) Start() {
	err:=r.Ed.Subscribe()
	if err!=nil{
		log.ErrorF(err.Msg(),err.Group(),err.Key())
	}
}
/*
func (w *Router) LogEvent(e ev.Event) error{
	if _, ok := w.Mgo.Cols["db_log"]; ok {
		res := w.Mgo.Cols["db_log"].Create(nil, e)
		if res.Status != status.DBStatus.Ok {
			return errors.New(fmt.Sprintf("% : %",res.ErrorCode,res.Message))
		}
	}
	return nil
}
*/
func (r *Router) ExitsEvent(event string) bool{
	if 	Map_PublisherContains(r.Pub,event){
		return true
	}
	return false
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
func (r *Router) GetServiceName(eventName string) (string,error){
	return utils.GetServiceName(eventName)
}
func (r *Router) GetServiceMethod(eventName string) (string,error){
	return utils.GetServiceMethod(eventName)
}