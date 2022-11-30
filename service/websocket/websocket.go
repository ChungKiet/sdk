package websocket

import (
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/base/event"
	r "github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/config/vault"
	ed "github.com/goonma/sdk/eventdriven"
	j "github.com/goonma/sdk/jwt"
	"github.com/goonma/sdk/log"
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
	key          string
	Srv          *echo.Echo
	Hub          *Hub
	//event driven subscriber, only 1
	Ev ed.EventDriven
	//micro client
	Client map[string]*micro.MicroClient
	// Redis
	rd r.CacheHelper
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
		TokenLookup:   "query:token",
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

	// init Subscriber
	var err_s *e.Error
	check, err_p := w.config.CheckPathExist("websocket/" + service_name + "/sub/kafka")
	if err_p != nil {
		log.ErrorF(err_p.Msg(), w.config.GetServiceName())
	}
	if check {
		err_s = w.Ev.InitialSubscriber(
			w.config, fmt.Sprintf("%s/%s/%s", "websocket", service_name, "sub/kafka"),
			service_name,
			subCallbackfn,
			nil)
		if err_s != nil {
			log.ErrorF(err_s.Msg(), err_s.Group(), err_s.Key())
		}
	}

	check, err_p = w.config.CheckPathExist("websocket/" + service_name + "/pub/kafka")
	if err_p != nil {
		log.ErrorF(err_p.Msg(), w.config.GetServiceName())
	}
	if check {
		err_s = w.Ev.InitialPublisher(w.config, fmt.Sprintf("%s/%s/%s", "websocket", service_name, "pub/kafka"), service_name)
		if err_s != nil {
			log.ErrorF(err_s.Msg(), err_s.Group(), err_s.Key())
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

}

func (w *Websocket) Start() {
	if w.host == "" || w.port == "" {
		log.Error("Please Initial before make new server")
		os.Exit(0)
	}
	fmt.Printf("HTTP Server start at: %s:%s\n", w.host, w.port)
	go w.Hub.run()
	// Start server
	go func() {
		if err := w.Srv.Start(":" + w.port); err != nil && err != http.ErrServerClosed {
			w.Srv.Logger.Fatal("shutting down the server")
		}
	}()

	err := w.Ev.Subscribe()
	if err != nil {
		log.ErrorF(err.Msg(), err.Group(), err.Key())
	}
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