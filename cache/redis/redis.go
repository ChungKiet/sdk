package redis

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"
	"fmt"
	redis "github.com/redis/go-redis/v9"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/utils"
)

type RedisHelper struct {
	Client *redis.Client
}
//standalone
func InitRedis(addr, password string, db_index int) (*redis.Client, *e.Error) {
	rdbclient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db_index, // use default DB
		MaxRetries: 3,
		//PoolTimeout:  2 * time.Minute,
		PoolSize: 1000,
		//IdleTimeout:  10 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		//IdleCheckFrequency: time.Second * 5,
	})
	_, err := rdbclient.Ping(context.Background()).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "INIT REDIS")
	}
	return rdbclient, nil
}
//
func InitRedisSentinel(addr_arr_str, master_name,password string, db_index int)(*redis.Client, *e.Error) {
	addr_arr:=utils.Explode(addr_arr_str,",")
	if len(addr_arr)==0{
		return nil, e.New("Have no redis host", "REDIS", "INIT SENTINE REDIS")
	}
	fmt.Println("Redis sentinel host:",addr_arr)
	fmt.Println("Redis sentinel mastername:",master_name)
	redisdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    master_name,
		SentinelAddrs: addr_arr,
		Password: password, // no password set
		SentinelPassword: password,
		DB:       db_index, // use default DB
		PoolSize: 1000,
		//IdleTimeout:  10 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		//IdleCheckFrequency: time.Second * 5,
	})
	_,err:=redisdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "INIT SENTINE REDIS")
	}
	return redisdb, nil
}
func (h *RedisHelper) Close() *e.Error {
	if h.Client!=nil{
		err := h.Client.Close()
		if err != nil {
			return e.New(err.Error(), "REDIS", "CLOSE REDIS")
		}
	}
	return nil
}

func (h *RedisHelper) Exists(key string) (bool,*e.Error) {
	if h.Client==nil{
		return false,e.New("Redis Client is null", "REDIS", "EXIST REDIS")
	}
	indicator, err := h.Client.Exists(context.Background(),key).Result()
	if err != nil {
		return false,e.New(err.Error(), "REDIS", "EXIST REDIS")
	}
	if indicator <= 0 {
		return false,nil
	}
	return true,nil
}

func (h *RedisHelper) Get(key string) (interface{},*e.Error) {
	if h.Client==nil{
		return nil,e.New("Redis Client is null", "REDIS", "REDIS GET")
	}
	data, err := h.Client.Get(context.Background(),key).Result()
	if err != nil {
		return nil,e.New(err.Error(), "REDIS", "GET_KEY")
	}
	var value interface{}
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil,e.New(err.Error(), "REDIS", "GET_KEY")
	}
	return value,nil
}
//return new value of key after increase old value
func (h *RedisHelper) IncreaseInt(key string,value int) (int,*e.Error) {
	if h.Client==nil{
		return 0,e.New("Redis Client is null", "REDIS", "REDIS GET")
	}
	res:=0
	//
	ctx:=context.Background()
	err := h.Client.Watch(ctx,func(tx *redis.Tx) error {
		n, err := tx.Get(ctx,key).Int()
		if err != nil && err != redis.Nil {
			return err
		}
	
		_, err = tx.Pipelined(ctx,func(pipe redis.Pipeliner) error {
			res=n+value
			pipe.Set(ctx,key, res,  time.Duration(300)* time.Second)
			return nil
		})
		return err
	}, key)
	//
	if err != nil {
		return 0,e.New(err.Error(), "REDIS", "IncreaseInt")
	}
	return res,nil
}
func (h *RedisHelper) GetInterface(key string, value interface{}) (interface{}, *e.Error) {
	if h.Client==nil{
		return nil,e.New("Redis Client is null", "REDIS", "REDIS GetInterface")
	}
	data, err := h.Client.Get(context.Background(),key).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS")
	}

	typeValue := reflect.TypeOf(value)
	kind := typeValue.Kind()

	var outData interface{}
	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		outData = reflect.New(typeValue).Interface()
	default:
		outData = reflect.Zero(typeValue).Interface()
	}
	err = json.Unmarshal([]byte(data), &outData)
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS")
	}

	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		outDataValue := reflect.ValueOf(outData)

		if reflect.Indirect(reflect.ValueOf(outDataValue)).IsZero() {
			return nil, e.New(errors.New("Get redis nil result").Error(), "REDIS", "GET INTERFACE REDIS")
		}
		if outDataValue.IsZero() {
			return outDataValue.Interface(), nil
		}
		return outDataValue.Elem().Interface(), nil
	}
	var outValue interface{} = outData
	if reflect.TypeOf(outData).ConvertibleTo(typeValue) {
		outValueConverted := reflect.ValueOf(outData).Convert(typeValue)
		outValue = outValueConverted.Interface()
	}
	return outValue, nil
}

func (h *RedisHelper) Set(key string, value interface{}, expiration time.Duration) *e.Error {
	if h.Client==nil{
		return e.New("Redis Client is null", "REDIS", "REDIS Set")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS")
	}

	_, err = h.Client.Set(context.Background(),key, string(data), expiration).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS")
	}
	return nil
}

func (h *RedisHelper) SetNX(key string, value interface{}, expiration time.Duration) (bool, *e.Error) {
	if h.Client==nil{
		return false,e.New("Redis Client is null", "REDIS", "REDIS SetNX")
	}
	var isSuccessful bool
	data, err := json.Marshal(value)
	if err != nil {
		return false, e.New(err.Error(), "REDIS", "SETNX REDIS")
	}

	isSuccessful, err = h.Client.SetNX(context.Background(),key, string(data), expiration).Result()
	if err != nil {
		return false, e.New(err.Error(), "REDIS", "SETNX REDIS")
	}
	return isSuccessful, nil
}

func (h *RedisHelper) Del(key string) *e.Error {
	if h.Client==nil{
		return e.New("Redis Client is null", "REDIS", "REDIS Del")
	}
	_, err := h.Client.Del(context.Background(),key).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "DEL REDIS")
	}
	return nil
}

func (h *RedisHelper) Expire(key string, expiration time.Duration) *e.Error {
	if h.Client==nil{
		return e.New("Redis Client is null", "REDIS", "REDIS Expire")
	}
	_, err := h.Client.Expire(context.Background(),key, expiration).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "EXPIRE REDIS")
	}
	return nil
}

func (h *RedisHelper) DelMulti(keys ...string) *e.Error {
	if h.Client==nil{
		return e.New("Redis Client is null", "REDIS", "REDIS DelMulti")
	}
	var err error
	pipeline := h.Client.TxPipeline()
	pipeline.Del(context.Background(),keys...)
	_, err = pipeline.Exec(context.Background())
	return e.New(err.Error(), "REDIS", "DEL MULTIPLE REDIS")
}

func (h *RedisHelper) GetKeysByPattern(pattern string) ([]string, uint64, *e.Error) {
	if h.Client==nil{
		return nil,0,e.New("Redis Client is null", "REDIS", "REDIS GetKeysByPattern")
	}
	var (
		keys   []string
		cursor uint64 = 0
		limit  int64  = 100
	)
	for {
		var temp_keys []string
		temp_keys, cursor, err := h.Client.Scan(context.Background(),cursor, pattern, limit).Result()
		if err != nil {
			return nil, 0, e.New(err.Error(), "REDIS", "GET KEYS REDIS")
		}

		keys = append(keys, temp_keys...)

		if cursor == 0 {
			break
		}
	}

	return keys, cursor, nil
}

func (h *RedisHelper) RenameKey(oldkey, newkey string) *e.Error {
	if h.Client==nil{
		return e.New("Redis Client is null", "REDIS", "REDIS RenameKey")
	}
	var err error
	_, err = h.Client.Rename(context.Background(),oldkey, newkey).Result()
	return e.New(err.Error(), "REDIS", "RENAME KEY REDIS")
}

func (h *RedisHelper) GetType(key string) (string, *e.Error) {
	if h.Client==nil{
		return "",e.New("Redis Client is null", "REDIS", "REDIS GetType")
	}
	typeK, err := h.Client.Type(context.Background(),key).Result()
	if err != nil {
		return "", e.New(err.Error(), "REDIS", "GET TYPE REDIS")
	}
	return typeK, nil
}

