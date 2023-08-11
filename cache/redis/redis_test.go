package redis

import (
	"fmt"
	"testing"
)

func mustNewRedisSentinel() *RedisHelper {
	c, err := InitRedis("10.148.0.42:6379", "EmB5mq6aTNTHJpffpkXRjvAcqrFtBf", 1)
	if err != nil {
		panic(err)
	}
	return &RedisHelper{
		Client: c,
	}
}

func TestRedisHelper_MGet(t *testing.T) {
	c := mustNewRedisSentinel()

	data, err := c.MGet("VCBVND", "HPGVND")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)
}

func TestRedisHelper_HMGet(t *testing.T) {
	c := mustNewRedisSentinel()

	data, err := c.HMGet("stocks", "HPGVND")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)
}

func TestRedisHelper_HGetAll(t *testing.T) {
	c := mustNewRedisSentinel()

	data, err := c.HGetAll("stocks")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(data)
}
