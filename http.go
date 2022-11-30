package main
import(
	httpServer "github.com/goonma/sdk/service/http"
	"github.com/goonma/sdk/utils"
	"github.com/labstack/echo/v4"
	"net/http"
	"github.com/goonma/sdk/jwt"
	"fmt"
)
type HTTP struct{
	httpServer.HTTPServer
}
func (sv *HTTP)hello(c echo.Context) error {
	sv.GetToken(c)
	return c.String(http.StatusOK, "Hello, World!")
}
func authenmethod(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
//white list route without check jwt 
var whitelist_routes []string
//ACL: role_id,routes permision 
var acl map[string]interface{}
func main(){
	key:=`jM+KpUEtvIRE+t8TsH*XgTqnGktYsmaA7%%N&zWudMFJv5eY)yXvY(3!qkcxkK`
	tk,_:=jwt.GenerateJWTToken(key,"1","abc","","Login", 1,600)
	fmt.Println(tk)
	claim,err:=jwt.VerifyJWTToken(key,"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiI2MzY3ZDY2ODQ4ZWNiZDQ4YmY1Y2NhMWQiLCJyb2xlaWQiOjEsInVzZXJuYW1lIjoiIiwiZW1haWwiOiJuZ3V5ZW5kdWNsb25nLmljdEBnbWFpbC5jb20iLCJ0eXBlIjoiMSIsImlzcyI6Ikdvb25tYS5jb20iLCJleHAiOjE2Njk4MjgyNTd9.zuWLtpavxJsrgmpO8wWleFC0vPUF3JDvn7FKw9OvSueoHBArEf8DZkwmIuOCiWwvwW8-_Q3wUue_JXxSHxpelw")
	fmt.Printf("%+v",claim)
	fmt.Println(err)
	//whitelist route
	whitelist_routes=[]string{"/hello"}
	//**********ACL*************
	acl=utils.Dictionary()
	//ACL-role 1
	role_1:=utils.DictionaryBool()
	role_1["/login"]=true
	role_1["/user/resetpassword"]=true
	acl["1"]=role_1
	//**********END ACL*************
	var w HTTP
	//
	w.Initial("api",nil,false,nil,whitelist_routes,acl)
	w.SetPathKey("path/")
	//add route
	w.Srv.GET("/hello", w.hello)
	w.Srv.POST("/authenmethod", authenmethod)
	//
	w.Start()
}