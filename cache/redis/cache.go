package redis

import (
	//"context"
	"time"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"
	"github.com/goonma/sdk/config/vault"
	"fmt"
)

type CacheHelper interface {
	Exists(key string) (bool,*e.Error)
	Get(key string) (interface{},*e.Error)
	GetInterface(key string, value interface{}) (interface{}, *e.Error)
	Set(key string, value interface{}, expiration time.Duration) *e.Error
	Del(key string) *e.Error
	Expire(key string, expiration time.Duration) *e.Error
	DelMulti(keys ...string) *e.Error
	GetKeysByPattern(pattern string) ([]string, uint64, *e.Error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, *e.Error)
	RenameKey(oldKey, newKey string) *e.Error
	GetType(key string) (string, *e.Error)
	Close() *e.Error
	IncreaseInt(key string,value int) (int,*e.Error)
}

// CacheOption represents cache option
type CacheOption struct {
	Key   string
	Value interface{}
}

func NewCacheHelper(vault *vault.Vault) (CacheHelper,*e.Error) {
	//
	config:=GetConfig(vault,"cache/redis")
	addrs:=utils.Explode(config["HOST"],",")
	password:=config["PASSWORD"]
	db:=config["DB"]
	db_index:=utils.StringToInt(db)
	if db_index<0{//default db
		db_index=0
	}
	log.Info(fmt.Sprintf("Initialing Redis: %s-%s",config["HOST"]),db_index)
	if len(addrs)==0{
		return nil,e.New("Redis host not found","REDIS","NewCacheHelper")
	}else{
		if addrs[0]==""{
			return nil,e.New("Redis host not found","REDIS","NewCacheHelper")
		}
	}
	//
	if len(addrs) > 1 {
		clusterClient, err := InitRedisCluster(addrs,password)
		if err != nil {
			return nil,err
		}
		fmt.Sprintf("Redis: %s %s",config["HOST"]," connected")
		return &ClusterRedisHelper{
			Client: clusterClient,
		},nil
	}
	client, err := InitRedis(addrs[0], password, db_index)
	if err != nil {
		return nil,err
	}
	log.Info(fmt.Sprintf("Redis: %s %s",config["HOST"]," connected"))
	return &RedisHelper{
		Client: client,
	},nil
}
func NewCacheHelperWithConfig(addrs []string,password string, db_index int) (CacheHelper,*e.Error) {
	//
	if len(addrs) > 1 {
		clusterClient, err := InitRedisCluster(addrs,password)
		if err != nil {
			return nil,err
		}
		
		return &ClusterRedisHelper{
			Client: clusterClient,
		},nil
	}
	client, err := InitRedis(addrs[0], password, db_index)
	if err != nil {
		return nil,err
	}
	
	return &RedisHelper{
		Client: client,
	},nil
}