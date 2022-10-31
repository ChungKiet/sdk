package main
import (
	"github.com/goonma/sdk/service/micro"
	"fmt"
	"context"
)
func main(){
	client,err:=micro.NewMicroClient("127.0.0.1")
	fmt.Println(err)
	data := "test"
	req_data_sv := map[string]interface{}{
		"data":data,
	}
	//fmt.Printf("Client GetInceptionTimeStamp Request: %+v\r\n",req_data_sv)
	ctx:=context.Background()
	resp_data_svc, err_c := client.CallService(ctx, "test", req_data_sv)
	fmt.Println(err_c)
	fmt.Println(resp_data_svc)
}