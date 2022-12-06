package main

import (
	micropb "github.com/goonma/sdk/base/pb/micro"
	"github.com/goonma/sdk/service/grpc"
	"github.com/goonma/sdk/service/micro"
	"github.com/goonma/sdk/base/event"
	resp "github.com/goonma/sdk/base/response"
	"context"
	"fmt"
)

type Server struct {
	micro.Micro
}

func (sv *Server) Handler(
	ctx context.Context,
	req *micropb.MicroRequest) (*micropb.MicroResponse, error) {
	fmt.Printf("%+v",req)
	return  resp.RaiseSuccess("Success", "test"), nil
}

func main() {
	var grpc grpc.GRPCServer
	//micro service
	var micro Server
	//initial grpc server
	grpc.Initial("micro.local.cluster.svc.goonma.demo")
	//
	micro.Initial(grpc.GetConfig(), nil,true)
	//
	micropb.RegisterMicroServiceServer(grpc.GetService(), &micro)
	ev:=event.Event{
		EventName:"pre-order",
		EventData: map[string]interface{}{
			"a":"1",
		},
	}
	err:=micro.PushEvent(ev)
	fmt.Println(err)
	//grpc.Start()
}
