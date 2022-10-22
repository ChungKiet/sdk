package health
import(
	"net/http"
	//"github.com/goonma/sdk/log"
	"io"
	"time"
	"fmt"
)
type HealthServer struct{
	readiness bool  
	liveness bool
	server  *http.Server
}
func NewHealth(port string) *http.Server{
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s","0.0.0.0",port),
		Handler: &HealthServer{
			readiness:false,
			liveness:true,
		},
		ReadTimeout: 5*time.Second,
	}
	return server
}
func (hs *HealthServer)SetReadiness(v bool){
	hs.readiness=v
}
func (hs *HealthServer)SetLiveness(v bool){
	hs.liveness=v
}
func (hs *HealthServer)Readiness(w http.ResponseWriter, r *http.Request) {
	if hs.readiness {
		w.WriteHeader(200)
		io.WriteString(w, "Ok")
	}else{
		w.WriteHeader(500)
		io.WriteString(w, "Fail")
	}
}
func (hs *HealthServer)Liveness(w http.ResponseWriter, r *http.Request) {
	if hs.liveness {
		w.WriteHeader(200)
		io.WriteString(w, "Ok")
	}else{
		w.WriteHeader(500)
		io.WriteString(w, "Fail")
	}
}
func (hs *HealthServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "URL: "+r.URL.String())
}

