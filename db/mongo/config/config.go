package config
import (
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/utils"
	"strings"
	"fmt"
)
func GetConfig(vault *vault.Vault,config_path string) map[string]string{
	config_path=strings.TrimSuffix(config_path,"/")
	m:=utils.DictionaryString()
	m["HOST"]=vault.ReadVAR(fmt.Sprintf("%s/HOST",config_path))
	m["DB"]=vault.ReadVAR(fmt.Sprintf("%s/DB",config_path))
	m["AUTHDB"]=vault.ReadVAR(fmt.Sprintf("%s/AUTHDB",config_path))
	m["USERNAME"]=vault.ReadVAR(fmt.Sprintf("%s/USERNAME",config_path))
	m["PASSWORD"]=vault.ReadVAR(fmt.Sprintf("%s/PASSWORD",config_path))
	m["USE_SSL"]=vault.ReadVAR(fmt.Sprintf("%s/USE_SSL",config_path))
	m["REPLICASET"]=vault.ReadVAR(fmt.Sprintf("%s/REPLICASET",config_path))
	m["READ_CONCERN"]=vault.ReadVAR(fmt.Sprintf("%s/READ_CONCERN",config_path))
	//ACK when: 0: journal; <0: N instances were written, all: all intances
	m["WRITE_CONCERN"]=vault.ReadVAR(fmt.Sprintf("%s/WRITE_CONCERN",config_path))
	m["COLLECTION"]=vault.ReadVAR(fmt.Sprintf("%s/COLLECTION",config_path))
	return m
}
func MergeConfig(global,local map[string]string) map[string]string{
	m:=utils.DictionaryString()
	if utils.Map_contains(global,"HOST") || utils.Map_contains(local,"HOST"){
		if utils.Map_contains(local,"HOST") && local["HOST"]!=""{
			m["HOST"]=local["HOST"]
		}else{
			m["HOST"]=global["HOST"]
		}
	}
	if utils.Map_contains(global,"DB") || utils.Map_contains(local,"DB"){
		if utils.Map_contains(local,"DB") &&  local["DB"]!=""{
			m["DB"]=local["DB"]
		}else{
			m["DB"]=global["DB"]
		}
	}
	if utils.Map_contains(global,"AUTHDB") || utils.Map_contains(local,"AUTHDB"){
		if utils.Map_contains(local,"AUTHDB") && local["AUTHDB"]!=""{
			m["AUTHDB"]=local["AUTHDB"]
		}else{
			m["AUTHDB"]=global["AUTHDB"]
		}
	}
	if utils.Map_contains(global,"USERNAME") || utils.Map_contains(local,"USERNAME"){
		if utils.Map_contains(local,"USERNAME") && local["USERNAME"]!=""{
			m["USERNAME"]=local["USERNAME"]
		}else{
			m["USERNAME"]=global["USERNAME"]
		}
	}
	if utils.Map_contains(global,"PASSWORD") || utils.Map_contains(local,"PASSWORD"){
		if utils.Map_contains(local,"PASSWORD") && local["PASSWORD"]!=""{
			m["PASSWORD"]=local["PASSWORD"]
		}else{
			m["PASSWORD"]=global["PASSWORD"]
		}
	}
	if utils.Map_contains(global,"USE_SSL") || utils.Map_contains(local,"USE_SSL"){
		if utils.Map_contains(local,"USE_SSL") && local["USE_SSL"]!=""{
			m["USE_SSL"]=local["USE_SSL"]
		}else{
			m["USE_SSL"]=global["USE_SSL"]
		}
	}
	if utils.Map_contains(global,"REPLICASET") || utils.Map_contains(local,"REPLICASET"){
		if utils.Map_contains(local,"REPLICASET") && local["REPLICASET"]!=""{
			m["REPLICASET"]=local["REPLICASET"]
		}else{
			m["REPLICASET"]=global["REPLICASET"]
		}
	}
	if utils.Map_contains(global,"READ_CONCERN") || utils.Map_contains(local,"READ_CONCERN"){
		if utils.Map_contains(local,"READ_CONCERN") && local["READ_CONCERN"]!=""{
			m["READ_CONCERN"]=local["READ_CONCERN"]
		}else{
			m["READ_CONCERN"]=global["READ_CONCERN"]
		}
	}
	if utils.Map_contains(global,"WRITE_CONCERN") || utils.Map_contains(local,"WRITE_CONCERN"){
		if utils.Map_contains(local,"WRITE_CONCERN") && local["WRITE_CONCERN"]!=""{
			m["WRITE_CONCERN"]=local["WRITE_CONCERN"]
		}else{
			m["WRITE_CONCERN"]=global["WRITE_CONCERN"]
		}
	}
	return m
}