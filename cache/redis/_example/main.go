package main

import (
	"context"
	"fmt"
	"time"

	redis "github.com/goonma/sdk/db/redis"
)

func main() {

	addr := []string{
		"127.0.0.1:6379",
	}
	type ABC struct {
		A string
		B string
		C string
	}
	abc := redis.NewCacheHelper(addr)
	abc.Set(
		context.Background(), "abc", "zcxv", time.Duration(1*time.Minute),
	)

	b := &ABC{
		A: "A",
		B: "B",
		C: "C",
	}
	err := abc.Set(
		context.Background(), "bcdd", b, time.Duration(1*time.Minute),
	)
	err = abc.Set(
		context.Background(), "bcddd", b, time.Duration(1*time.Minute),
	)
	err = abc.Set(
		context.Background(), "bcdddd", b, time.Duration(1*time.Minute),
	)
	err = abc.Set(
		context.Background(), "bcddddn", b, time.Duration(1*time.Minute),
	)
	fmt.Println(err)
	d := &ABC{}
	var a string
	err = abc.Get(context.Background(), "bcdd", d)
	fmt.Println(a, b, d, err)

	keys, nCursor, err := abc.GetKeysByPattern(context.Background(), "*bc*")
	fmt.Println(nCursor, keys)
}
