package http
import (
	"github.com/labstack/echo/v4"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/config/vault"
)
type HTTPServer struct {
	host        string
	port        string
	servicename string
	//key-value store management
	config *vault.Vault
	//http server
	Srv *echo.Echo
	//db map[string]dbconnection
	Mgo  db.MongoDB
	//event
	Ed ed.EventDriven
	//micro client
	Client map[string]*MicroClient
}
/*
args[0]: model list
args[1]: not exist || exist && true then initial publisher, else don't implement publisher 
args[2]: micro client list map[string]string (name - endpoint address)
*/

func (sv *HTTPServer) Initial(service_name string,args...interface{}){
	//get ENV
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	log.Initial(service_name)
	//initial Server configuration
	var config vault.Vault
	sv.config= &config
	sv.config.Initial(service_name)
	//get config from key-value store
	http_port_env:=grpcSRV.config.ReadVAR("micro/general/HTTP_PORT")
	if micro_port_env!=""{
		sv.port=http_port_env
	}
	//
	if os.Getenv("HTTP_HOST")==""{
		sv.host="0.0.0.0"
	}else{
		sv.host=os.Getenv("HTTP_HOST")
	}
	if os.Getenv("HTTP_PORT")!=""{
		sv.port=os.Getenv("HTTP_PORT")
	}else if http.host=="" {
		sv.port="8080"		
	}
	//set service name
	sv.servicename=service_name
	//ReInitial Destination for Logger
	if log.LogMode()!=2{// not in local, locall just output log to std
		log_dest:=sv.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest=="kafka"{
			config_map:=kafka.GetConfig(grpcSRV.config,"logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//publisher
	if len(args)>1{
		// load event_routing(event_name, bus_name) table/ event_mapping db
		pub_path:=fmt.Sprintf("%s/%s/%s","api",sv.servicename,"pub/kafka")
		check,err_p:=sv.config.CheckPathExist(path)
		if err_p!=nil{
			log.ErrorF(err_p.Msg(),sv.servicename,"Initial")
		}
		if check{
			event_list:=sv.config.ListItemByPath(pub_path)
			sv.Pub=make(map[string]*ed.EventDriven)
			for _,event:=range event_list{
				if !Map_PublisherContains(sv.Pub,event) && event!="general"{
					sv.Pub[event]=&ed.EventDriven{}
					//r.Pub[event].SetNoUpdatePublishTime(true)
					err:=sv.Pub[event].InitialPublisherWithGlobal(sv.config,fmt.Sprintf("%s/%s",pub_path,event),"Router",event)
					if err!=nil{
						log.ErrorF(err.Msg(),sv.servicename,"Initial")
					}
				}
			}
		}
	}else{
		c,err:=utils.ItoBool(args[1])
		if err!=nil{
			log.Warn("Convert Iterface to Bool :"+err.Error(),"MICRO","HOST_NAME")
		}
		if utils.Type(args[1])=="bool" && c{
			err_p:=micro.Ed.InitialPublisher(micro.Config,"eventbus/kafka",micro.Id)
			if err_p!=nil{
				log.ErrorF(err_p.Msg(),err_p.Group(),err_p.Key())
			}
		}
	}
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
	//new server
	sv.Srv=echo.New()
}

//start
func (sv *HTTPServer) Start(){
	if sv.host=="" || sv.port==""{
		log.Error("Please Initial before make new server")
		os.Exit(0)
	}
	sv.Srv.Logger.Fatal(sv.Srv.Start(":1323"))
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