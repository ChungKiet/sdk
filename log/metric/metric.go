package metric

import (
	"os"
	"github.com/goonma/sdk/pubsub/kafka"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	//"fmt"
	"encoding/json"
	"time"
	"errors"
)


type Metric struct {
	id string
	host string
	service     string
	env         string 
	pub  kafka.Publisher
}
var metric Metric
func Initial(service_name string,config_map map[string]string) error {
	//
	//get ENV
	err_env := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err_env!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	env:=os.Getenv("ENV")//2: local,1: development,0:product
	metric.service = service_name
	hostname, _ := os.Hostname()
	metric.host=hostname
	metric.env  = env
	err:=metric.pub.InitialManual(config_map,"Metrics")
	if err!=nil{
		return errors.New(err.Msg())
	}
	return nil
}
func Push(metric_name string,t1,t2 time.Time) {
	t:=t2.Sub(t1).Seconds()
	metric_data:=map[string]interface{}{
		"id":uuid.New().String(),
		"host": metric.host,
		"service": metric.service,
		"metric": metric_name,
		"t": t,
		"env": metric.env,
	}
	data, err := json.Marshal(metric_data)
    if err == nil {
		metric.pub.Publish(data)
    }
}
