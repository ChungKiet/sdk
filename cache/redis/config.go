package redis
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
	m["PASSWORD"]=vault.ReadVAR(fmt.Sprintf("%s/PASSWORD",config_path))
	return m
}