package worker

import (
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/base/event"
	ev "github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/log/metric"

	//"github.com/goonma/sdk/pubsub/kafka"
	r "github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/db/mongo"
	"github.com/goonma/sdk/pubsub/kafka"
	"github.com/goonma/sdk/service/micro"
	"github.com/goonma/sdk/utils"

	//"github.com/goonma/sdk/db/mongo/status"
	//e "github.com/goonma/sdk/base/error"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	//"errors"
)

type Worker struct {
	//worker name
	worker_name string
	//key-value store management
	config *vault.Vault
	//event driven subscriber, only 1
	Sub ed.EventDriven
	//event driven for multi publisher
	Pub map[string]*ed.EventDriven
	//db map[string]dbconnection
	Mgo db.MongoDB
	//log publisher
	log_pub ed.EventDriven
	//redis
	Rd r.CacheHelper
	//micro client
	Client map[string]*micro.MicroClient
	//default init subscriber log for push item processs success to kafka_item_sucess
	uninit_subscriber_log bool
	//detaul init redis cache
	init_redis bool
}

// initial n worker
func (w *Worker) Initial(worker_name string, callbackfn event.ConsumeFn, args ...interface{}) {
	//get ENV
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err != nil {
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err != nil {
			panic(err)
		}
	}
	log.Initial(worker_name)
	w.worker_name = worker_name
	//initial Server configuration
	var config vault.Vault
	w.config = &config
	w.config.Initial(fmt.Sprintf("%s/%s", "worker", worker_name))
	//
	//ReInitial Destination for Logger
	if log.LogMode() != 2 { // not in local, locall just output log to std
		log_dest := w.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest == "kafka" {
			config_map := kafka.GetConfig(w.config, "logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//init metric
	config_map := kafka.GetConfig(w.config, "metric/kafka")
	err_m := metric.Initial(worker_name, config_map)
	if err_m != nil {
		log.Warn(err_m.Error(), "InitMetrics")
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
	//
	//initial Event subscriber
	var err_s *e.Error
	//default alway init subscriber_log for push item consumer process success to log_item_sucess, log item fail push from router base on TTL
	if w.uninit_subscriber_log { // not default case
		err_s = w.Sub.InitialSubscriber(
			w.config, fmt.Sprintf("%s/%s/%s", "worker", worker_name, "sub/kafka"),
			worker_name,
			callbackfn,
			nil)
		fmt.Println("===Disable Log for item success===")
	} else { //default case
		err_s = w.Sub.InitialSubscriber(
			w.config, fmt.Sprintf("%s/%s/%s", "worker", worker_name, "sub/kafka"),
			worker_name,
			callbackfn,
			w.LogEvent)
		fmt.Println("===Init Log for item success===")
	}
	if err_s != nil {
		log.ErrorF(err_s.Msg(), err_s.Group(), err_s.Key())
	}
	//initial Event publisher (gRPC micro service publisher)
	fmt.Println("===Init publisher for kafka main_bus===")
	//initial Event published
	check, err_p := w.config.CheckPathExist("worker/" + worker_name + "/pub/kafka")
	if err_p != nil {
		log.ErrorF(err_p.Msg(), worker_name, "Initial")
	}
	fmt.Println("Check:", check)
	w.Pub = make(map[string]*ed.EventDriven)
	if check { //custom publisher, list event
		event_list := w.config.ListItemByPath("worker/" + worker_name + "/pub/kafka")
		for _, event := range event_list {
			if !Map_PublisherContains(w.Pub, event) && event != "general" {
				w.Pub[event] = &ed.EventDriven{}
				//micro.Pub[event].SetNoUpdatePublishTime(true)
				err := w.Pub[event].InitialPublisherWithGlobal(w.config, fmt.Sprintf("worker/%s/%s/%s", worker_name, "pub/kafka", event), worker_name, event)
				if err != nil {
					log.ErrorF(err.Msg(), worker_name, "Initial")
				}
			}
		}
	} else { //use main bus
		w.Pub["main"] = &ed.EventDriven{}
		err_p := w.Pub["main"].InitialPublisher(w.config, "eventbus/kafka", worker_name)
		if err_p != nil {
			log.ErrorF(err_p.Msg(), err_p.Group(), err_p.Key())
		}
	}
	//for Repush Event to main_bus when Consumer Process fail
	fmt.Println("===Init publisher(of consumer) for repush item fail to mainbus===")
	w.Sub.SetPublisherForSubscriber()
	//init consumed Log
	//initial DB args[0] => mongodb
	if len(args) > 1 {
		if args[1] != nil {
			models, err := utils.ItoDictionary(args[1])
			if err != nil {
				log.ErrorF(err.Error(), "WOKER", "INITIAL_CONVERTION_MODEL")
			}
			err_init := w.Mgo.Initial(w.config, models)
			if err_init != nil {
				log.Warn(err_init.Msg(), err_init.Group(), err_init.Key())
			}
		}
	}
	//micro client call service
	if len(args) > 0 {
		if args[0] != nil {
			remote_services, err := utils.ItoDictionaryS(args[0])
			if err != nil {
				log.ErrorF(err.Error(), "WORKER", "INITIAL_CONVERTION_MODEL")
			} else {
				w.InitialMicroClient(remote_services, false)
			}
		}
	}
	//defaul alway init subscriber_log for push item processed sucess to log_item_sucess. for item process fail, push by router base on TTL
	if !w.uninit_subscriber_log {
		w.InitConsumedLog()
	}

}
func (w *Worker) InitConsumedLog() {
	//initial publisher for streaming consumed Item to DB for easy tracking
	w.log_pub.SetNoUpdatePublishTime(true)
	err := w.log_pub.InitialPublisher(w.config, fmt.Sprintf("%s/%s/%s", "worker", w.worker_name, "consumed_log/kafka"), w.worker_name, "Consummer Logger")
	if err != nil {
		log.ErrorF(err.Msg(), err.Group(), err.Key())
	}
	//
}

// not need to call if want to use SubscriberLog
// SubscriberLog => for push item success to log_item_success, for log_item_fail push from router
func (w *Worker) SetNoInitSubscriberLog(i bool) {
	w.uninit_subscriber_log = i
}

func (w *Worker) SetInitRedis(i bool) {
	w.init_redis = i
}
func (w *Worker) AutoSubscriberACK(v bool) {
	w.Sub.AutoSubscriberACK(v)
}
// start consumers
func (w *Worker) Start() {
	err := w.Sub.Subscribe()
	if err != nil {
		log.ErrorF(err.Msg(), err.Group(), err.Key())
	}
}

func (w *Worker) LogEvent(e ev.Event) error {
	err := w.log_pub.Publish(e)
	if err != nil {
		log.Error(err.Msg(), err.Group(), err.Key(), e)
	}
	return nil
}
func (w *Worker) GetServiceName(eventName string) (string, error) {
	return utils.GetServiceName(eventName)
}
func (w *Worker) GetServiceMethod(eventName string) (string, error) {
	return utils.GetServiceMethod(eventName)
}
func (w *Worker) InitialMicroClient(remote_services map[string]string, initConnection bool) {
	//
	w.Client = make(map[string]*micro.MicroClient)
	//
	var err *e.Error
	for k, v := range remote_services {
		if k != "" {
			if initConnection {
				w.Client[k], err = micro.NewMicroClient(v)
				if err != nil {
					log.ErrorF(err.Msg(), err.Group(), err.Key())
				} else {
					log.Info(fmt.Sprintf("Micro Client: %s->%s %s", k, v, " Initial success"))
				}
			} else {
				w.Client[k], err = micro.NewMicroClientWithoutConnection(v)
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

func (w *Worker) GetDB(col_name string) (*mongo.Collection, *e.Error) {
	if w.Mgo.Cols[col_name] != nil {
		return w.Mgo.Cols[col_name], nil
	} else {
		return nil, e.New(fmt.Sprintf(" DB not initial: %s", col_name), "Worker", "GetDB")
	}
}
func (w *Worker) GetClient(client_name string) (*micro.MicroClient, *e.Error) {
	if w.Client[client_name] != nil {
		return w.Client[client_name], nil
	} else {
		return nil, e.New(fmt.Sprintf(" Micro Client not initial: %s", client_name), "Worker", "GetClient")
	}
}
func (w *Worker) GetConfig() *vault.Vault {
	return w.config
}
func (w *Worker) Clean() {
	w.Sub.Clean()
	w.Mgo.Clean()
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
