package redis
import (
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/utils"
	"strings"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)
func GetConfig(vault *vault.Vault,config_path string) map[string]string{
	m:=utils.DictionaryString()
	err := godotenv.Load(os.ExpandEnv("/config/.env"))
	if err!=nil{
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err!=nil{
			config_path=strings.TrimSuffix(config_path,"/")
			m["HOST"]=vault.ReadVAR(fmt.Sprintf("%s/HOST",config_path))
			m["DB"]=vault.ReadVAR(fmt.Sprintf("%s/DB",config_path))
			m["PASSWORD"]=vault.ReadVAR(fmt.Sprintf("%s/PASSWORD",config_path))
			m["TYPE"]=vault.ReadVAR(fmt.Sprintf("%s/TYPE",config_path))
			m["MASTER_NAME"]=vault.ReadVAR(fmt.Sprintf("%s/MASTER_NAME",config_path))
			return m
		}
	}
	//
	m["HOST"]=os.Getenv("HOST")
	m["DB"]=os.Getenv("DB")
	m["PASSWORD"]=os.Getenv("PASSWORD")
	m["TYPE"]=os.Getenv("TYPE")
	m["MASTER_NAME"]=os.Getenv("MASTER_NAME")
	return m
}