package grpc

import (
	"net"
	"os"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/jwt"
	//"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/pubsub/kafka"
	"fmt"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"os/signal"
	"syscall"
	"context"
	"errors"
)

type GRPCServer struct {
	host        string
	port        string
	servicename string
	//key-value store management
	config *vault.Vault
	//grpc server
	service *grpc.Server
	two_FA_Key  string
	token_Key  string
}

func (grpcSRV *GRPCServer) Initial(service_name string){
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
	grpcSRV.config= &config
	grpcSRV.config.Initial(service_name)
	//get config from key-value store
	micro_port_env:=grpcSRV.config.ReadVAR("micro/general/MICRO_PORT")
	if micro_port_env!=""{
		grpcSRV.port=micro_port_env
	}
	//
	if os.Getenv("MICRO_HOST")==""{
		grpcSRV.host="0.0.0.0"
	}else{
		grpcSRV.host=os.Getenv("MICRO_HOST")
	}
	if os.Getenv("MICRO_PORT")!=""{
		grpcSRV.port=os.Getenv("MICRO_PORT")
	}else if grpcSRV.host=="" {
		grpcSRV.port="30000"		
	}
	//set service name
	grpcSRV.servicename=service_name
	//ReInitial Destination for Logger
	if log.LogMode()!=2{// not in local, locall just output log to std
		log_dest:=grpcSRV.config.ReadVAR("logger/general/LOG_DEST")
		if log_dest=="kafka"{
			config_map:=kafka.GetConfig(grpcSRV.config,"logger/kafka")
			log.SetDestKafka(config_map)
		}
	}
	//new grpc server
	maxMsgSize := 1024 * 1024 * 1024 //1GB
	//read 2FA Key for verify token
	grpcSRV.two_FA_Key=grpcSRV.config.ReadVAR("key/2fa/KEY")
	grpcSRV.token_Key=grpcSRV.config.ReadVAR("key/api/KEY")
	//
	grpcSRV.service= grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize), 
		grpc.MaxSendMsgSize(maxMsgSize),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(grpcSRV.authFunc)),//middleware verify authen
	)
}
/*
Start gRPC server with IP:Port from Initial step
*/
func (grpcSRV *GRPCServer) Start() {
	//check server
	if grpcSRV.host=="" || grpcSRV.port==""{
		log.Error("Please Initial before make new server")
		os.Exit(0)
	}
	errs_chan := make(chan error)
	stop_chan := make(chan os.Signal)
	// bind OS events to the signal channel
	signal.Notify(stop_chan, syscall.SIGTERM, syscall.SIGINT)
	//
	go grpcSRV.listen(errs_chan)
	//
	defer func() {
		grpcSRV.service.GracefulStop()
	 }()
	// block until either OS signal, or server fatal error
	select {
		case err := <-errs_chan:
			log.Error(err.Error(),"MICRO")
		case <-stop_chan:
 	}
	log.Warn("Service shutdown","MICRO")

} 
func (grpcSRV *GRPCServer) listen(errs chan error) {
	grpcAddr:= net.JoinHostPort(grpcSRV.host, grpcSRV.port)
	listener, err:= net.Listen("tcp", grpcAddr)
	if err != nil {
		log.ErrorF(err.Error())
	}
	//
	log.Info(fmt.Sprintf("gRPC service started: %s - %s",grpcSRV.servicename,grpcAddr))
	//
	//healthgrpc.RegisterHealthSever(server, hs)
	//reflection.Register(server)
	errs <- grpcSRV.service.Serve(listener)
}
func (grpcSRV *GRPCServer)GetService() *grpc.Server{
	return grpcSRV.service
}
func (grpcSRV *GRPCServer)GetConfig() *vault.Vault{
	return grpcSRV.config
}

func (grpcSRV *GRPCServer)authFunc(ctx context.Context) (context.Context, error) {
	fmt.Println(grpcSRV.servicename)
	//ignore check token
	if os.Getenv("IGNORE_TOKEN")=="true"{
		return ctx,nil
	}
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	//fmt.Println(token)
	if err != nil {
		return nil, err
	}
	_, err_v := jwt.VerifyJWTToken(grpcSRV.token_Key,token)
	if err_v != nil {
		return nil, err_v
	}
	//verify permision base on service name + method name
	method_route,res:=grpc.Method(ctx)
	if !res{
		return nil,errors.New("Access Deny")
	}
	if method_route==""{
		return nil,errors.New("Method is empty")
	}
	arr:=utils.Explode(method_route,"/")
	if len(arr)!=2{
		return nil,errors.New("Method invalid")
	}
	method:=arr[2]
	//fmt.Println(method)
	//

	//grpc_ctxtags.Extract(ctx).Set("auth.sub", userClaimFromToken(tokenInfo))
	// WARNING: in production define your own type to avoid context collisions
	//newCtx := context.WithValue(ctx, "tokenInfo", tokenInfo)
	//return newCtx, nil*/
	return ctx,nil
}
