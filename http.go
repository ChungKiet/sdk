package main
import(
	"github.com/goonma/sdk/service/http"
)
type HTTP struct{
	http.HTTPServer
}
func main(){
	//
	var w HTTP
	//
	w.Initial("api")
	//
	w.Start()
}