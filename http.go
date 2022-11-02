package main
import(
	httpServer "github.com/goonma/sdk/service/http"
	"github.com/goonma/sdk/utils"
	"github.com/labstack/echo/v4"
	"net/http"
)
type HTTP struct{
	httpServer.HTTPServer
}
func hello(c echo.Context) error {
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
	//whitelist route
	whitelist_routes=[]string{"/hello"}
	//**********ACL*************
	//ACL-role 1
	role_1:=utils.DictionaryBool()
	role_1["/login"]=true
	role_1["/user/resetpassword"]=true
	acl["1"]=role_1
	//**********END ACL*************
	var w HTTP
	//
	w.Initial("api",nil,false,nil,whitelist_routes,acl)
	//add route
	w.Srv.GET("/hello", hello)
	w.Srv.POST("/authenmethod", authenmethod)
	//
	w.Start()
}