package main 
import (
	worker "github.com/goonma/sdk/service/worker/publisher"
	//ev "github.com/goonma/sdk/eventdriven"
	//"github.com/goonma/sdk/log"
	//"github.com/goonma/sdk/pubsub/kafka"	
	//"os"
	//"github.com/joho/godotenv"
	//"context"
	//"errors"
	"fmt"
)
type Worker struct{
	worker.Worker
}
func main(){
	//
	var w Worker
	
	w.Initial("worker-demo")//paste function process worker here
	w.wkProcess()
	
}

func (wk *Worker)wkProcess() error{
	fmt.Println("start worker publisher")
	return nil
}
