package websocket

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/base/event"
	ev "github.com/goonma/sdk/base/event"
	r "github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	j "github.com/goonma/sdk/jwt"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/log/metric"
	"github.com/goonma/sdk/pubsub/kafka"
	"github.com/goonma/sdk/service/micro"
	"github.com/goonma/sdk/utils"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Websocket struct {
	host         string
	port         string
	service_name string
	config       *vault.Vault

	key string

	Srv *echo.Echo

	Hub *Hub

	//Ws handler func

	//event driven subscriber, only 1
	Sub ed.EventDriven
	//retry delete Uid in redis publisher
	//redis_pub ed.EventDriven
	//log publisher
	log_pub ed.EventDriven
	//event driven for multi publisher
	Pub map[string]*ed.EventDriven
	//micro client
	Client map[string]*micro.MicroClient

	// Redis
	rd r.CacheHelper

	//default don't subscriber log for push item process success to kafka_item_success
	init_subscriber_log bool
	//defaul don't check redis Uid
	init_check_uid bool
}

func (w *Websocket) Initial(service_name string, wsHandleFunc echo.HandlerFunc, subCallbackfn event.ConsumeFn, args ...interface{}) {
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err != nil {
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err != nil {
			panic(err)
		}
	}
	log.Initial(service_name)
	w.service_name = service_name

	w.Hub = newHub()

	//initial Server configuration
	var config vault.Vault
	w.config = &config
	w.config.Initial(service_name)

	//read secret key for generate JWT
	w.key = w.config.ReadVAR("key/api/KEY")

	if os.Getenv("HTTP_HOST") == "" {
		w.host = "0.0.0.0"
	} else {
		w.host = os.Getenv("HTTP_HOST")
	}
	if os.Getenv("HTTP_PORT") != "" {
		w.port = os.Getenv("HTTP_PORT")
	} else if w.host == "" {
		w.port = "8080"
	}

	w.Srv = echo.New()
	w.Srv.HideBanner = true
	w.Srv.HidePort = true

	w.Srv.Use(middleware.Logger())
	w.Srv.Use(middleware.Recover())
	w.Srv.GET("/ws", wsHandleFunc)

	config_jwt := middleware.JWTConfig{
		Claims:        &j.CustomClaims{},
		SigningKey:    []byte(w.key),
		SigningMethod: jwt.SigningMethodHS256.Name,
		TokenLookup:   "header:" + echo.HeaderAuthorization + ", query:token",
		AuthScheme:    "Bearer",
		ParseTokenFunc: func(token string, c echo.Context) (interface{}, error) {
			if os.Getenv("IGNORE_TOKEN") == "true" {
				return nil, nil
			}
			claims_info, err := j.VerifyJWTToken(w.key, token)
			if err != nil {
				return nil, err
			}
			return claims_info, nil
		},
	}
	w.Srv.Use(middleware.JWTWithConfig(config_jwt))

	// Init redis
	redis, err_r := r.NewCacheHelper(w.config)
	if err_r != nil {
		log.ErrorF(err_r.Msg())
	}
	w.rd = redis

	//ReInitial Destination for Logger
	if log.LogMode() != 2 { // not in local, local just output log to std
		log_dest := w.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest == "kafka" {
			config_map := kafka.GetConfig(w.config, "logger/kafka")
			log.SetDestKafka(config_map)
		}
	}

	//init metric
	config_map := kafka.GetConfig(w.config, "metric/kafka")
	err_m := metric.Initial(service_name, config_map)
	if err_m != nil {
		log.Warn(err_m.Error(), "InitMetrics")
	}

	// init Subscriber
	var err_s *e.Error
	if w.init_subscriber_log {
		err_s = w.Sub.InitialSubscriber(
			w.config, fmt.Sprintf("%s/%s/%s", "websocket", service_name, "sub/kafka"),
			service_name,
			subCallbackfn,
			w.LogEvent)
	} else {
		err_s = w.Sub.InitialSubscriber(
			w.config, fmt.Sprintf("%s/%s/%s", "websocket", service_name, "sub/kafka"),
			service_name,
			subCallbackfn,
			nil)
	}
	if err_s != nil {
		log.ErrorF(err_s.Msg(), err_s.Group(), err_s.Key())
	}

	check, err_p := w.config.CheckPathExist("websocket/" + service_name + "/pub/kafka")
	if err_p != nil {
		log.ErrorF(err_p.Msg(), w.config.GetServiceName())
	}
	// Init publisher
	w.Pub = make(map[string]*ed.EventDriven)
	if check {
		event_list := w.config.ListItemByPath("websocket/" + service_name + "/pub/kafka")
		for _, event := range event_list {
			if !Map_PublisherContains(w.Pub, event) && event != "general" {
				w.Pub[event] = &ed.EventDriven{}
				//micro.Pub[event].SetNoUpdatePublishTime(true)
				err := w.Pub[event].InitialPublisherWithGlobal(w.config, fmt.Sprintf("%s/%s/%s", "websocket/"+service_name, "pub/kafka", event), w.config.GetServiceName(), event)
				if err != nil {
					log.ErrorF(err.Msg(), w.config.GetServiceName(), "Initial")
				}
			}
		}
	}
	// else {
	// 	w.Pub["main"] = &ed.EventDriven{}
	// 	err_p := w.Pub["main"].InitialPublisher(w.config, "eventbus/kafka", service_name)
	// 	if err_p != nil {
	// 		log.ErrorF(err_p.Msg(), err_p.Group(), err_p.Key())
	// 	}
	// }

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

	//consume log, reinit
	if w.init_subscriber_log {
		w.InitConsumedLog()
		fmt.Println("*****Initiation consumed log*****")
	} else {
		fmt.Println("*****Ignore consumed log*****")
	}
}

func (w *Websocket) Start() {
	if w.host == "" || w.port == "" {
		log.Error("Please Initial before make new server")
		os.Exit(0)
	}
	fmt.Printf("HTTP Server start at: %s:%s\n", w.host, w.port)

	// Start server
	go func() {
		if err := w.Srv.Start(":" + w.port); err != nil && err != http.ErrServerClosed {
			w.Srv.Logger.Fatal("shutting down the server")
		}
	}()

	err := w.Sub.Subscribe()
	if err != nil {
		log.ErrorF(err.Msg(), err.Group(), err.Key())
	}
}

func (w *Websocket) InitConsumedLog() {
	//initial publisher for streaming consumed Item to DB for easy tracking
	w.log_pub.SetNoUpdatePublishTime(true)
	err := w.log_pub.InitialPublisher(w.config, fmt.Sprintf("%s/%s/%s", "websocket", w.service_name, "consumed_log/kafka"), w.service_name, "Consummer Logger")
	if err != nil {
		log.ErrorF(err.Msg(), err.Group(), err.Key())
	}
}

func (w *Websocket) SetInitSubscriberLog(i bool) {
	w.init_subscriber_log = i
}
func (w *Websocket) SetInitCheckUID(i bool) {
	w.init_check_uid = i
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

func (w *Websocket) InitialMicroClient(remote_services map[string]string, initConnection bool) {
	w.Client = make(map[string]*micro.MicroClient)
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
}

func (w *Websocket) LogEvent(e ev.Event) error {
	err := w.log_pub.Publish(e)
	if err != nil {
		fmt.Println(err.Msg())
	}
	return nil
}
