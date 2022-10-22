package main
import (
	"github.com/goonma/sdk/db/redis"
	"fmt"
	//"time"
)
func main(){
	redis_ips:=[]string{"10.148.0.198:6379"}
	cache:=redis.NewCacheHelper(redis_ips,"Goonma@2022",0)
	cache.Set("key1",0,0)
	v,_:=cache.Get("key1")
	fmt.Println("Key1:",v)
	//cache.Set("key1", "value1",0)
	//for i:=0;i<60;i++{
		_,err:=cache.GetWatch("key1")
		fmt.Println("Key1 err:",err)
		//fmt.Println("Key1:",value)
		v,_=cache.Get("key1")
		fmt.Println("Key1:",v)
		//time.Sleep(1 * time.Second)
	//}
}