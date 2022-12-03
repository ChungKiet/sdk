package http
import (
	"github.com/labstack/echo/v4"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/db"
	"github.com/goonma/sdk/service/micro"
	"github.com/goonma/sdk/pubsub/kafka"
	ed "github.com/goonma/sdk/eventdriven"
	j "github.com/goonma/sdk/jwt"
	"github.com/goonma/sdk/config/vault"
	"github.com/labstack/echo/v4/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"os"
	"errors"
	"net/http"
	"os/signal"
	"context"
	"time"
	"fmt"
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
	Pub map[string]*ed.EventDriven
	//micro client
	Client map[string]*micro.MicroClient
	//serect key for JWT 
	key string
	path_key string
	//ACL
	Acl map[string]interface{}
}

/*
args[0]: model list
args[1]: not exist || exist && true then initial publisher, else don't implement publisher 
args[2]: micro client list map[string]string (name - endpoint address)
args[3]: route white list not check JWT
args[4]: ACL 
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
	http_port_env:=sv.config.ReadVAR("api/general/HTTP_PORT")
	if http_port_env!=""{
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
	}else if sv.host=="" {
		sv.port="8080"		
	}
	//set service name
	sv.servicename=service_name
	//ReInitial Destination for Logger
	if log.LogMode()!=2{// not in local, locall just output log to std
		log_dest:=sv.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest=="kafka"{
			config_map:=kafka.GetConfig(sv.config,"logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//read secret key for generate JWT
	sv.key=sv.config.ReadVAR("key/api/KEY")
	//publisher
	if len(args)>1{
		c,err:=utils.ItoBool(args[1])
		if err!=nil{
			log.Warn("Convert Iterface to Bool :"+err.Error(),"MICRO","HOST_NAME")
		}
		if utils.Type(args[1])=="bool" && c{
			// load event_routing(event_name, bus_name) table/ event_mapping db
			pub_path:=fmt.Sprintf("%s/%s/%s","api",sv.servicename,"pub/kafka")
			check,err_p:=sv.config.CheckPathExist(pub_path)
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
						err:=sv.Pub[event].InitialPublisherWithGlobal(sv.config,fmt.Sprintf("%s/%s",pub_path,event),sv.servicename,event)
						if err!=nil{
							log.ErrorF(err.Msg(),sv.servicename,"Initial")
						}
					}
				}
			}else{
				err_p:=sv.Pub["main"].InitialPublisher(sv.config,"eventbus/kafka",sv.servicename)
				if err_p!=nil{
					log.ErrorF(err_p.Msg(),err_p.Group(),err_p.Key())
				}				
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
				err_init:=sv.Mgo.Initial(sv.config,models)
				if err_init!=nil{
					log.Warn(err_init.Msg(),err_init.Group(),err_init.Key())
				}
			}
		}
	}
	//
	var routes_ignore_jwt []string
	var errh error
	if len(args)>3{
		routes_ignore_jwt,errh=utils.ItoSliceString(args[3])
		if errh!=nil{
			log.Warn(errh.Error(),"HttpServerInit","RoutesIgnoreJWT")
		}
	
	}
	if len(args)>4{
		var err error
		sv.Acl,err=utils.ItoDictionary(args[4])
		if err!=nil{
			log.Warn(errh.Error(),"HttpServerInit","ACL")
		}
	}
	//new server
	sv.Srv=echo.New()
	//
	sv.Srv.HideBanner = true
    sv.Srv.HidePort = true
 
    sv.Srv.Use(middleware.Logger())
    sv.Srv.Use(middleware.Recover())
	//middleware verify JWT
	//tk,err:=j.GenerateJWTToken(sv.key,"1","goonma","test@goonma.com","LOGIN",1,600000000000000)
	//if err!=nil{
	// log.ErrorF(errh.Error(),"HttpServerInit","JWT")
	//}
	//fmt.Println("err gen tk:%+v\r\n",err)
	//fmt.Println(tk)
	config_jwt:= middleware.JWTConfig{
		Claims:     &j.CustomClaims{},
		SigningKey: []byte(sv.key),
		SigningMethod: jwt.SigningMethodHS512.Name,
		//ContextKey:       "user",
		TokenLookup:      "header:" + echo.HeaderAuthorization+",query:token",
		//TokenLookupFuncs: nil,
		AuthScheme:       "Bearer",
		//KeyFunc:          nil,
		Skipper: func(c echo.Context) bool {
			if os.Getenv("IGNORE_TOKEN")=="true"{
				return true
			}
			if os.Getenv("ENV")=="local"{
				fmt.Println("Route: ",c.Request().URL.Path)
			}
			
			if utils.Contains(routes_ignore_jwt,c.Request().URL.Path) {
			  return true
			}
			return false
		},
		ParseTokenFunc: func(token string, c echo.Context) (interface{}, error) {
			if os.Getenv("IGNORE_TOKEN")=="true"{
				return nil,nil
			}
			if os.Getenv("ENV")=="local"{
				fmt.Println("KEY",sv.key,"Token: ",token)
			}
			
			Claims_info,err:=j.VerifyJWTToken(sv.key,token)
			if err!=nil{
				return nil,err 
			}
			if utils.MapI_contains(sv.Acl,utils.ItoString(Claims_info.RoleID)){
				m_acl,err:=utils.ItoDictionaryBool(sv.Acl[utils.ItoString(Claims_info.RoleID)])
				if err!=nil{
					return nil,err
				}
				if m_acl[c.Request().URL.Path] {
					return token,nil
				}
			}
			return token,nil
		},
	}
	/*
	group route
	g := e.Group("/admin", <your-middleware>)
	g.GET("/secured", <your-handler>) =>/admin/secured
	*/
	//disable CORS
	sv.Srv.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	sv.Srv.Use(middleware.JWTWithConfig(config_jwt))
}

//start
func (sv *HTTPServer) Start(){
	if sv.host=="" || sv.port==""{
		log.Error("Please Initial before make new server")
		os.Exit(0)
	}
	//
	fmt.Println(fmt.Sprintf("HTTP Server start at: %s:%s", sv.host, sv.port))
	// Start server
	go func() {
		if err := sv.Srv.Start(":"+sv.port); err != nil && err != http.ErrServerClosed {
			sv.Srv.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds. 
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sv.Srv.Shutdown(ctx); err != nil {
		sv.Srv.Logger.Fatal(err)
	}
}
func (sv *HTTPServer)SetPathKey(path string) error {
	key:=sv.config.ReadVAR(path)
	if key==""{
		return errors.New("_PATH_KEY_NOT_EXISTS_")
	}
	sv.key=path
	return nil
}
func (sv *HTTPServer)GetUserID(c echo.Context) (string,error) {
	token,err:=sv.GetToken(c)
	if err!=nil{
		return "",err 
	}
	Claims_info,err:=j.VerifyJWTToken(sv.key,token)
	if err!=nil{
		return "",err 
	}
	return Claims_info.UserID,nil
}
func (sv *HTTPServer)GetRoleID(c echo.Context) (int,error) {
	token,err:=sv.GetToken(c)
	if err!=nil{
		return 0,err 
	}
	Claims_info,err:=j.VerifyJWTToken(sv.key,token)
	if err!=nil{
		return 0,err 
	}
	return Claims_info.RoleID,nil
}
func (sv *HTTPServer)GetToken(c echo.Context) (string,error) {
	authen_str:=c.Request().Header.Get("Authorization")
	arr:=utils.Explode(authen_str," ")
	if len(arr)!=2{
		token:=c.QueryParam("token")
		if token==""{
			return "",errors.New("_JWT_INVALID_")
		}
		return token,nil
	}
	return arr[1],nil
}
func (sv *HTTPServer)Restricted(c echo.Context) error {
	//user := c.Get("user").(*jwt.Token)
	//claims := user.Claims.(*jwt.CustomClaims)
	//name := claims.Name
	//return c.String(http.StatusOK, "Welcome "+name+"!")
	return errors.New("Access Deny")
}
/*
func (sv *HTTPServer) JWTConfig() middleware.JWTConfig{
	config := middleware.JWTConfig{
		TokenLookup: "query:token",
		ParseTokenFunc: func(auth string, c echo.Context) (interface{}, error) {
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != "HS256" {
				return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return sv.key, nil
			}
			// claims are of type `jwt.MapClaims` when token is created with `jwt.Parse`
			token, err := jwt.Parse(auth, keyFunc)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, errors.New("invalid token")
			}
			return token, nil
		},
  	}
  	return config
}*/
func Map_PublisherContains(m map[string]*ed.EventDriven, item string) bool {
	if len(m)==0{
		return false
	}
	if _, ok := m[item]; ok {
		return true
	}
	return false
}