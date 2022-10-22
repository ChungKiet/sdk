package main 
import (
	"github.com/goonma/sdk/service/worker"
	//ev "github.com/goonma/sdk/eventdriven"
	//"github.com/goonma/sdk/log"
	//"github.com/goonma/sdk/pubsub/kafka"	
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/goonma/sdk/utils"
	"os"
	"github.com/joho/godotenv"
	//"context"
	//"errors"
	//"fmt"
)
type Worker struct{
	worker.Worker
}
func main(){
	//
	var w Worker
	//default worker don't initial publihser, if you want initial, set by bellow command
	w.InitPublisher(false)
	w.InitSubscriberLog(false)
	//no database
	w.Initial("woker1",w.wkProcess,RemoteServices())//paste function process worker here
	//with database
	//w.Initial("woker1",w.wkProcess,RemoteServices(),loadModel())//paste function process worker here
	//
	w.Start()
}
//process logic after consume Event from kafka
//iportant return error to errs channel, if error nil-> ACK will be send to kafka
//else ACK don't send, then Event will be loop forever
func (wk *Worker)wkProcess(message *message.Message) error{
	// wk.Mgo["col"].
	//fmt.Println("DATA:",message.UUID, string(message.Payload))
	// payload => Event struct
	//            EventName  => <service_name>-<method_name>
	//            EventData
	// marshal message.Payload -> Event
	// arr:=utils.Explode(EventName,"|")
	//service_name:=arr[0]
	//method_name:=""
	// if len(arr)>1 method_name=arr[1]
	// _,err:=service client=>CallService(ctx,service_name,method_name)
	// if err!=nil{
	//	errs<-err
	//}
	return nil
}

//declare service client, move to another file
func RemoteServices() map[string]string {
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err != nil {
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err != nil {
			panic(err)
		}
	}
	list := utils.DictionaryString()
	//declare remote services for stg && prd
	if os.Getenv("ENV") == "stg" || os.Getenv("ENV") == "prd"  || os.Getenv("ENV") == "dev"{
		//func
		list["fn.queue0"] = "micro.local.cluster.svc.treehouse.fn.queue0"
	} else { //declare remote services for local
		//func
		list["fn.queue0"] = "127.0.0.1:30001"
	}	
	return list
}
