package worker

import (
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/log/metric"
	ev "github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/base/event"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/pubsub/kafka"
	//"github.com/goonma/sdk/pubsub/kafka"
	"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/db/mongo"
	"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/service/micro"
	r "github.com/goonma/sdk/cache/redis"
	//"github.com/goonma/sdk/db/mongo/status"
	//e "github.com/goonma/sdk/base/error"
	"os"
	"github.com/joho/godotenv"
	"fmt"
	//"errors"
)
type Worker struct {
	//worker name
	worker_name string
	//key-value store management
	config *vault.Vault
	//event driven
	Ed ed.EventDriven
	//db map[string]dbconnection
	Mgo  db.MongoDB
	//redis
	Rd r.CacheHelper
	//log publisher
	log_pub ed.EventDriven
	//micro client
	Client map[string]*micro.MicroClient
	//default don't  publisher for repush item process fail
	init_publisher bool
	//default don't subscriber log for push item processs success to kafka_item_sucess
	init_subscriber_log bool
	//default don't update consume time + finish time
	no_update_processing_time bool
	//detaul init redis cache
	init_redis bool
}

//initial n worker
func (w *Worker) Initial(worker_name string,callbackfn event.ConsumeFn,args...interface{}) {
	//get ENV
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	log.Initial(worker_name)
	w.worker_name=worker_name
	//initial Server configuration
	var config vault.Vault
	w.config= &config
	w.config.Initial(fmt.Sprintf("%s/%s","worker",worker_name))
	//
	//ReInitial Destination for Logger
	if log.LogMode()!=2{// not in local, locall just output log to std
		log_dest:=w.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest=="kafka"{
			config_map:=kafka.GetConfig(w.config,"logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//init metric
	config_map:=kafka.GetConfig(w.config,"metric/kafka")
	err_m:=metric.Initial(worker_name,config_map)
	if err_m!=nil{
		log.Warn(err_m.Error(),"InitMetrics")
	}
	//
	//set noEvent => No inject ConsumeTime + Publish time
	if !w.no_update_processing_time{
		fmt.Println("*****Initiation no_inject true*****")
		w.Ed.SetNoEvent(true)
	}else{
		fmt.Println("*****Ignore no_inject false*****")
		w.Ed.SetNoEvent(false)
	}
	
	//initial Event subscriber
	//-for push item processed to kafka_item_succes, for kafka_item_fail push from router when TTL
	// default use subscriber_log it not set from worker implement
	var err_s *e.Error
	if w.init_subscriber_log{
		err_s=w.Ed.InitialSubscriber(
			w.config,fmt.Sprintf("%s/%s/%s","worker",worker_name,"sub/kafka"),
			worker_name,
			callbackfn,		
			w.LogEvent)
	}else{
		err_s=w.Ed.InitialSubscriber(
			w.config,fmt.Sprintf("%s/%s/%s","worker",worker_name,"sub/kafka"),
			worker_name,
			callbackfn,
			nil)
	}
	if err_s!=nil{
		log.ErrorF(err_s.Msg(),err_s.Group(),err_s.Key())
	}
	//initial Event published
	// for repush item to main bus when item process fail
	// default always init Publisher
	if w.init_publisher{
		err_p:=w.Ed.InitialPublisher(w.config,fmt.Sprintf("%s/%s/%s","worker",worker_name,"pub/kafka"),worker_name,"Event Bus")
		if err_p!=nil{
			log.ErrorF(err_p.Msg(),err_p.Group(),err_p.Key())
		}
		//for Repush Event when Consumer Process error
		w.Ed.SetPublisherForSubscriber()
	}
	//
	
	//init consumed Log
	//initial DB args[0] => mongodb
	if len(args)>1{
		if args[1]!=nil{
			models,err:=utils.ItoDictionary(args[1])
			if err!=nil{
				log.ErrorF(err.Error(),"WOKER","INITIAL_CONVERTION_MODEL")
			}
			err_init:=w.Mgo.Initial(w.config,models)
			if err_init!=nil{
				log.Warn(err_init.Msg(),err_init.Group(),err_init.Key())
			}
		}
	}
	//micro client call service
	if len(args)>0{
		if args[0]!=nil{
			remote_services,err:=utils.ItoDictionaryS(args[0])
			if err!=nil{
				log.ErrorF(err.Error(),"WORKER","INITIAL_CONVERTION_MODEL")
			}else{
				w.InitialMicroClient(remote_services,false)
			}
		}
	}
	//consume log, reinit
	if w.init_subscriber_log{
		w.InitConsumedLog()
		fmt.Println("*****Initiation consumed log*****")
	}else{
		fmt.Println("*****Ignore consumed log*****")
	}
	//init redis
	if w.init_redis {
		fmt.Println("===Init Redis===")
		redis, err_r := r.NewCacheHelper(w.config)
		if err_r != nil {
			log.ErrorF(err_r.Msg())
		}
		w.Rd = redis
	}
	
}
func (w *Worker) InitConsumedLog() {
	//initial publisher for streaming consumed Item to DB for easy tracking
	w.log_pub.SetNoUpdatePublishTime(true)
	err:=w.log_pub.InitialPublisher(w.config,fmt.Sprintf("%s/%s/%s","worker",w.worker_name,"consumed_log/kafka"),w.worker_name,"Consummer Logger")
	if err!=nil{
		log.ErrorF(err.Msg(),err.Group(),err.Key())
	}
	//
}
func (w *Worker) SetInitPublisher(i bool) {
	w.init_publisher=i
}
func (w *Worker) SetInitSubscriberLog(i bool) {
	w.init_subscriber_log=i
}
func (w *Worker) AutoSubscriberACK(v bool) {
	w.Ed.AutoSubscriberACK(v)
}
func (w *Worker) SetNoUpdateProcessingTime(i bool) {
	w.no_update_processing_time=i
}
//start consumers
func (w *Worker) Start() {
	err:=w.Ed.Subscribe()
	if err!=nil{
		log.ErrorF(err.Msg(),err.Group(),err.Key())
	}
}
//-push item processed to kafka_item_succes, for kafka_item_fail push from router when TTL
func (w *Worker) LogEvent(e ev.Event) error{
	err:=w.log_pub.Publish(e)
	if err!=nil{
		fmt.Println(err.Msg())
	}
	return nil
}
func (w *Worker) SetInitRedis(i bool) {
	w.init_redis = i
}
func (w *Worker) GetServiceName(eventName string) (string,error){
	return utils.GetServiceName(eventName)
}
func (w *Worker) GetServiceMethod(eventName string) (string,error){
	return utils.GetServiceMethod(eventName)
}
func (w *Worker) InitialMicroClient(remote_services map[string]string,initConnection bool){
	//
	w.Client=make(map[string]*micro.MicroClient)
	//
	var err *e.Error
	for k,v:=range remote_services{
		if k!=""{
			if initConnection{
				w.Client[k],err=micro.NewMicroClient(v)
				if err!=nil{
					log.ErrorF(err.Msg(),err.Group(),err.Key())
				}else{
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s",k,v," Initial success"))
				}
			}else{
				w.Client[k],err=micro.NewMicroClientWithoutConnection(v)
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

func (w *Worker)GetDB(col_name string) (*mongo.Collection,*e.Error){
	if w.Mgo.Cols[col_name]!=nil{
		return w.Mgo.Cols[col_name],nil
	}else{
		return nil,e.New(fmt.Sprintf(" DB not initial: %s",col_name),"Worker","GetDB")
	}
}
func (w *Worker)GetClient(client_name string) (*micro.MicroClient,*e.Error){
	if w.Client[client_name]!=nil{
		return w.Client[client_name],nil
	}else{
		return nil,e.New(fmt.Sprintf(" Micro Client not initial: %s",client_name),"Worker","GetClient")
	}
}
func (w *Worker)GetConfig() *vault.Vault{
	return w.config
}
func (w *Worker)Clean(){
	w.Ed.Clean()
	w.Mgo.Clean()
}
