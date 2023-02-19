package redis

import (
	//"context"
	"context"
	"encoding/json"
	"reflect"
	"time"

	//"fmt"
	//"errors"
	redis "github.com/go-redis/redis"
	e "github.com/goonma/sdk/base/error"
)

type ClusterRedisHelper struct {
	Client *redis.ClusterClient
}

// sharding
func InitRedisCluster(addrs []string, password string) (*redis.ClusterClient, *e.Error) {
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:      addrs,
		Password:   password,
		MaxRetries: 3,
		//PoolTimeout:  2 * time.Minute,
		PoolSize:           1000,
		IdleTimeout:        10 * time.Minute,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		IdleCheckFrequency: time.Second * 5,
	})
	_, err := clusterClient.Ping().Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "INIT REDIS CLUSTER")
	}
	return clusterClient, nil
}
func (h *ClusterRedisHelper) Exists(key string) (bool, *e.Error) {
	if h.Client == nil {
		return false, e.New("Redis Client is null", "REDIS", "REDIS Exists")
	}
	indicator, err := h.Client.Exists(key).Result()
	if err != nil {
		return false, e.New(err.Error(), "REDIS", "EXIST REDIS CLUSTER")
	}
	if indicator <= 0 {
		return false, nil
	}
	return true, nil
}

func (h *ClusterRedisHelper) Get(key string) (interface{}, *e.Error) {
	if h.Client == nil {
		return nil, e.New("Redis Client is null", "REDIS", "REDIS Get")
	}
	data, err := h.Client.Get(key).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET REDIS CLUSTER")
	}
	var value interface{}
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET REDIS CLUSTER")
	}
	return value, nil
}

func (h *ClusterRedisHelper) GetWithContext(ctx context.Context, key string) (interface{}, *e.Error) {

	if h.Client == nil {
		return nil, e.New("Redis Client is null", "REDIS", "REDIS Get")
	}
	data, err := h.Client.Get(key).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET REDIS CLUSTER")
	}
	var value interface{}
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET REDIS CLUSTER")
	}
	return value, nil
}

// return new value of key after increase old value
func (h *ClusterRedisHelper) IncreaseInt(key string, value int) (int, *e.Error) {
	if h.Client == nil {
		return 0, e.New("Redis Client is null", "REDIS", "REDIS GET")
	}
	res := 0
	//
	err := h.Client.Watch(func(tx *redis.Tx) error {
		n, err := tx.Get(key).Int()
		if err != nil && err != redis.Nil {
			return err
		}

		_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
			res = n + value
			pipe.Set(key, res, time.Duration(300)*time.Second)
			return nil
		})
		return err
	}, key)
	//
	if err != nil {
		return 0, e.New(err.Error(), "REDIS", "IncreaseInt")
	}
	return res, nil
}

func (h *ClusterRedisHelper) GetInterface(key string, value interface{}) (interface{}, *e.Error) {
	if h.Client == nil {
		return nil, e.New("Redis Client is null", "REDIS", "REDIS GetInterface")
	}
	var err error
	data, err := h.Client.Get(key).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS CLUSTER")
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
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS CLUSTER")
	}

	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		return reflect.ValueOf(outData).Elem().Interface(), nil
	}
	var outValue interface{} = outData
	if reflect.TypeOf(outData).ConvertibleTo(typeValue) {
		outValueConverted := reflect.ValueOf(outData).Convert(typeValue)
		outValue = outValueConverted.Interface()
	}
	return outValue, nil
}

func (h *ClusterRedisHelper) GetInterfaceWithContext(ctx context.Context, key string, value interface{}) (interface{}, *e.Error) {
	if h.Client == nil {
		return nil, e.New("Redis Client is null", "REDIS", "REDIS GetInterface")
	}
	var err error
	data, err := h.Client.Get(key).Result()
	if err != nil {
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS CLUSTER")
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
		return nil, e.New(err.Error(), "REDIS", "GET INTERFACE REDIS CLUSTER")
	}

	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		return reflect.ValueOf(outData).Elem().Interface(), nil
	}
	var outValue interface{} = outData
	if reflect.TypeOf(outData).ConvertibleTo(typeValue) {
		outValueConverted := reflect.ValueOf(outData).Convert(typeValue)
		outValue = outValueConverted.Interface()
	}
	return outValue, nil
}

func (h *ClusterRedisHelper) Set(key string, value interface{}, expiration time.Duration) *e.Error {
	if h.Client == nil {
		return e.New("Redis Client is null", "REDIS", "REDIS Set")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS CLUSTER")
	}
	_, err = h.Client.Set(key, string(data), expiration).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS CLUSTER")
	}
	return nil
}

func (h *ClusterRedisHelper) SetWithContext(ctx context.Context, key string, value interface{}, expiration time.Duration) *e.Error {
	if h.Client == nil {
		return e.New("Redis Client is null", "REDIS", "REDIS Set")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS CLUSTER")
	}
	_, err = h.Client.Set(key, string(data), expiration).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "SET REDIS CLUSTER")
	}
	return nil
}

func (h *ClusterRedisHelper) SetNX(key string, value interface{}, expiration time.Duration) (bool, *e.Error) {
	if h.Client == nil {
		return false, e.New("Redis Client is null", "REDIS", "REDIS SetNX")
	}
	var isSuccessful bool
	data, err := json.Marshal(value)
	if err != nil {
		return false, e.New(err.Error(), "REDIS", "SETNX REDIS CLUSTER")
	}
	isSuccessful, err = h.Client.SetNX(key, string(data), expiration).Result()
	if err != nil {
		return false, e.New(err.Error(), "REDIS", "SETNX REDIS CLUSTER")
	}
	return isSuccessful, nil
}

func (h *ClusterRedisHelper) Del(key string) *e.Error {
	if h.Client == nil {
		return e.New("Redis Client is null", "REDIS", "REDIS Del")
	}
	_, err := h.Client.Del(key).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "DEL REDIS CLUSTER")
	}
	return nil
}

func (h *ClusterRedisHelper) DelMulti(keys ...string) *e.Error {
	//??
	var err *e.Error
	return err
}

func (h *ClusterRedisHelper) Expire(key string, expiration time.Duration) *e.Error {
	if h.Client == nil {
		return e.New("Redis Client is null", "REDIS", "REDIS Expire")
	}
	_, err := h.Client.Expire(key, expiration).Result()
	if err != nil {
		return e.New(err.Error(), "REDIS", "EXPIRE REDIS CLUSTER")
	}
	return nil
}

func (h *ClusterRedisHelper) GetKeysByPattern(pattern string) ([]string, uint64, *e.Error) {
	//???
	var err *e.Error
	var keys []string
	return keys, 0, err
}

func (h *ClusterRedisHelper) RenameKey(oldkey, newkey string) *e.Error {
	///?
	return nil
}

func (h *ClusterRedisHelper) GetType(key string) (string, *e.Error) {
	if h.Client == nil {
		return "", e.New("Redis Client is null", "REDIS", "REDIS GetType")
	}
	typeK, err := h.Client.Type(key).Result()
	if err != nil {
		return "", e.New(err.Error(), "REDIS", "GET TYPE REDIS CLUSTER")
	}
	return typeK, nil
}
func (h *ClusterRedisHelper) Close() *e.Error {
	//?
	return nil
}

func (h *ClusterRedisHelper) IncreaseMinWithPattern(pattern string, value int) (string, *e.Error) {

	return "", nil
}
