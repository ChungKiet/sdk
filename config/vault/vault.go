package vault

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"github.com/hashicorp/vault/api"
	"github.com/goonma/sdk/log"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/utils"
	"strings"
	"os"
	"github.com/joho/godotenv"
)

type Vault struct {
	host      string
	port      string
	username  string
	password  string
	token     string
	root_path string
	servicename string
	client    *api.Client
	
}
func (v *Vault) Initial(service_name string, args ...string) {
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			panic(err)
		}
	}
	//check KEY_STORE
	dir:="secret/data/"
	if os.Getenv("KEY_STORE_HOST")==""{
		log.ErrorF("KEY_STORE_HOST is empty","ENV_ERROR","KEY_STORE_HOST")
	}
	if os.Getenv("KEY_STORE_USER")==""{
		log.ErrorF("KEY_STORE_USER is empty","ENV_ERROR","KEY_STORE_USER")
	}
	if os.Getenv("KEY_STORE_PASSWORD")==""{
		log.ErrorF("KEY_STORE_USER is empty","ENV_ERROR","KEY_STORE_PASSWORD")
	}
	if os.Getenv("KEY_STORE_DIR")!=""{
		dir=os.Getenv("KEY_STORE_DIR")
	}
	host:=os.Getenv("KEY_STORE_HOST")
	port:=os.Getenv("KEY_STORE_PORT")
	user:=os.Getenv("KEY_STORE_USER")
	pass:=os.Getenv("KEY_STORE_PASSWORD")
	
	//
	if host == "" {
		log.ErrorF("VAULT Host is empty", "VAULT_ERROR", "VAULT_HOST_EMPTY")
	}
	if port == "" {
		port = "8200"
	}
	if user == "" {
		log.ErrorF("VAULT User is empty", "VAULT_ERROR", "VAULT_USER_EMPTY")
	}
	if pass == "" {
		log.ErrorF("VAULT Password is empty", "VAULT_ERROR", "VAULT_PASSWORD_EMPTY")
	}
	v.host = host
	v.port = port
	v.username = user
	v.password = pass
	v.servicename=service_name
	//
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	vaultAddr := fmt.Sprintf("%s:%s",v.host, v.port)
	client, err := api.NewClient(&api.Config{Address: vaultAddr, HttpClient: httpClient})
	if err != nil {
		log.ErrorF(err.Error(), "VAULT_ERROR")
	}
	options := map[string]interface{}{
		"password": v.password,
	}
	path := fmt.Sprintf("auth/userpass/login/%s", v.username)
	secret, err := client.Logical().Write(path, options)
	if err != nil {
		log.ErrorF(err.Error(), "VAULT_ERROR")
	}
	v.token = secret.Auth.ClientToken
	client.SetToken(v.token)
	v.client = client
	//
	
	//
	if len(args) > 0 {
		v.root_path = args[0]
	} else {
		v.root_path = dir
		//v.root_path = "kv/data/"
	}
	if utils.Right(v.root_path,1)!="/"{
		v.root_path=v.root_path+"/"
	}
	log.Info(fmt.Sprintf("%s %s %s %s: %s","VAULT server ",v.host," root path: ",v.root_path," connected"),"VAULT","INITIATION")
}
func (v *Vault) InitialByToken(host, port, token string, service_name string, args ...string) {
	if host == "" {
		log.ErrorF("VAULT Host not found", "VAULT_ERROR", "VAULT_HOST_EMPTY")
	}
	if port == "" {
		port = "8200"
	}
	v.host = host
	v.port = port
	v.token = token
	v.servicename=service_name
	//
	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	vaultAddr := net.JoinHostPort(v.host, v.port)
	client, err := api.NewClient(&api.Config{Address: vaultAddr, HttpClient: httpClient})
	if err != nil {
		log.ErrorF(err.Error(), "VAULT_ERROR")
	}
	client.SetToken(token)
	v.client = client
	//
	if len(args) > 0 {
		v.root_path = args[0]
	} else {
		v.root_path = "secret/data/"
		//v.root_path = "kv/data/"
	}
	if utils.Right(v.root_path,1)!="/"{
		v.root_path=v.root_path+"/"
	}
	log.Info(fmt.Sprintf("%s %s %s %s: %s","VAULT server ",v.host," root path: ",v.root_path," connected"),"VAULT","INITIATION")
}
func (v *Vault) ReadVAR(path string) string {
	if path==""{
		log.Error("ENV path is empty", "VAULT_ERROR")
	}
	//fmt.Println(v.root_path)
	arr:=utils.Explode(path,"/")
	var_name:=arr[len(arr)-1]
	folder:=strings.Join(arr[:len(arr)-1], "/")
	//fmt.Println(fmt.Sprintf("%s/%s",v.root_path,folder))
	//data, err := v.client.Logical().Read("secret/metadata/data/worker/woker1/sub/kafka")
	data, err := v.client.Logical().Read(fmt.Sprintf("%s%s",v.root_path,folder))
	//data, err := v.client.Logical().List("secret/metadata/data/worker/kafka-to-alert-mgt/sub/kafka")
	//fmt.Printf("%+v\r\n",data)
	//fmt.Println(err)
	if err != nil {
		log.Warn(err.Error(), "VAULT_ERROR")
		return ""
	}
	if data==nil{
		log.Warn(fmt.Sprintf("%s not found",path), "VAULT_ERROR")
		return ""
	}
	for k,v:=range data.Data{
		if k==var_name{
			return utils.ItoString(v)
		}
	}
	return ""
}
func (v *Vault) CheckPathExist(path string) (bool,*e.Error) {
	if path==""{
		return false,nil
	}
	data, err := v.client.Logical().List(fmt.Sprintf("%s%s",v.root_path,path))
	if err != nil {
		return false,e.New(err.Error(),"VAULT","CHECK PATH")
	}
	if data==nil{
		return false,nil
	}
	if data.Data==nil{
		return false,nil
	}
	return true,nil
}
func (v *Vault) CheckItemExist(path string) (bool,*e.Error) {
	if path==""{
		return false,nil
	}
	arr:=utils.Explode(path,"/")
	item:=arr[len(arr)-1]
	path=strings.Join(arr[:len(arr)-1], "/")
	data, err := v.client.Logical().List(fmt.Sprintf("%s%s",v.root_path,path))
	if err != nil {
		return false,e.New(err.Error(),"VAULT","CHECK PATH")
	}
	if data==nil{
		return false,nil
	}
	if data.Data==nil{
		return false,nil
	}
	m,err:=utils.ItoSliceString(data.Data["keys"])
	if err!=nil{
		return false,nil
	}
	for _,k:=range m{
		if k==item{
			return true,nil
		}
	}
	return false,nil
}
func (v *Vault) ListItemByPath(path string) []string {
	if path==""{
		log.Error("ENV path is empty", "VAULT_ERROR")
	}
	data, err := v.client.Logical().List(fmt.Sprintf("%s%s",v.root_path,path))
	if err != nil {
		log.Error(err.Error(), "VAULT_ERROR")
		return nil
	}
	if data==nil{
		log.Error(fmt.Sprintf("%s not found",path), "VAULT_ERROR")
		return nil
	}
	if data.Data==nil{
		log.Error(fmt.Sprintf("%s not found",path), "VAULT_ERROR")
		return nil
	}
	if !utils.MapI_contains(data.Data,"keys"){
		log.Error(fmt.Sprintf("%s not found",path), "VAULT_ERROR")
		return nil
	}
	res,err:=utils.ItoSliceString(data.Data["keys"])
	if err!=nil{
		log.Error(err.Error(), "VAULT_ERROR","CONVERT_INTERFACE_TO_ARRAY")
		return nil
	}
	return res
}
func (v *Vault) SetServiceName(service_name string){
	v.servicename=service_name
}
func (v *Vault) GetServiceName() string{
	return v.servicename
}
func (v *Vault)GetClient() *api.Client{
	return v.client
}