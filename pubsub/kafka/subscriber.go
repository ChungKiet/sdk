package kafka
import(
	"github.com/goonma/sdk/utils"
	//"github.com/goonma/sdk/utils/transform"
	//"github.com/goonma/sdk/health"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/log"
	ev "github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/config/vault"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/goonma/sdk/cache/redis"
	"github.com/Shopify/sarama"
	"os"
	//"errors"
	"fmt"
	"context"
	"encoding/json"
	"time"
	"math"
	"os/signal"
	"syscall"
	//"errors"
)
//type ConsumeFn = func(messages <-chan *message.Message)
type Subscriber struct {
	id string
	subscriber *kafka.Subscriber
	topic string
	ProcessFn event.ConsumeFn
	RePushEventFn event.RePushFn
	RetryDeleteRedisFn	event.RetryDeleteRedisPushFn
	Redis redis.CacheHelper
	logConsumeFn event.WriteLogConsumeFn
	num_consumer int
	no_ack bool
	no_inject bool
	no_cache bool	
	//
	config map[string]string
}

//Initial Publisher
func (sub *Subscriber) Initial(vault *vault.Vault,config_path string,worker_name string,callbackfn event.ConsumeFn,logConsume event.WriteLogConsumeFn) *e.Error{
	log.Info("Initialing Kafka subscriber...","KAFKA")
	//
	hostname, err_h := os.Hostname()
	if err_h != nil {
		log.Warn(fmt.Sprintf("Error get Hostname: %s",err_h.Error()))
	}
	//prd, stg no resend
	if os.Getenv("ENV")=="prd" ||  os.Getenv("ENV")=="stg" || os.Getenv("ENV") == "dev"{
		sub.SetNoAck(false)
	}else{//local ENv resend message
		sub.SetNoAck(true)
	}
	sub.id=worker_name+"_"+hostname
	//
	
	//
	config_map:=GetConfig(vault,config_path)
	sub.config=config_map
	//conf:=NewConfig(config_map)
	//fmt.Printf("%+v",config_map)
	brokers_str:=config_map["BROKERS"]
	topic:=config_map["TOPIC"]
	consumer_group:=config_map["CONSUMER_GROUP"]
	num_consumer:=utils.ItoInt(config_map["NUM_CONSUMER"])
	if num_consumer==math.MinInt32{
		return e.New("Event Bus Number of Consumer must be number","KAFKA","CONSUMER")
	}
	sub.num_consumer=num_consumer
	
	if brokers_str==""{
		//log.ErrorF("Event Bus Brokers not found","KAFKA","KAFKA_CONSUMER_BROKER")
		return e.New("Event Bus Brokers not found","KAFKA","CONSUMER")
	}
	if topic==""{
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_CONSUMER_TOPIC")
		return e.New("Event Bus Topic not found","KAFKA","CONSUMER")
	}
	if consumer_group==""{
		//log.ErrorF("Event Bus Consumer Group not found","KAFKA","CONSUMER")
		return e.New("Event Bus Consumer Group not found","KAFKA","CONSUMER")
	}
	
	sub.topic=topic
	//
	var err error
	brokers:=utils.Explode(brokers_str,",")
	//conf:=kafka.DefaultSaramaSubscriberConfig()
	conf:=NewConsumerConfig(config_map)
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	//sarama.OffsetOldest get from last offset not yet commit
	//sarama.OffsetNewest  ignore all mesage just get new message after consumer start
	sub.subscriber, err = kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:                brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
			ConsumerGroup:        consumer_group,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)
	if err!=nil{
		if log.LogMode()!=0{
			data_cfg_str,err_c:=utils.MapToJSONString(config_map)
			if err_c!=nil{
				//log.ErrorF(err.Error(),"KAFKA")
				return e.New(err_c.Error()+": "+data_cfg_str,"KAFKA","CONSUMER")
			}
			return e.New(fmt.Sprintf("%s: %s",err.Error(),data_cfg_str),"KAFKA","CONSUMER")
		}else{
			return e.New(err.Error(),"KAFKA","CONSUMER")
		}
	}
	log.Info(fmt.Sprintf("%s %s: %s","Kafka consumer brokers: ",brokers_str+" | "+topic," connected"),"KAFKA","CONSUMER")
	sub.ProcessFn=callbackfn
	if logConsume!=nil{
		fmt.Println("===========Initiation Processed Item Log======")
		sub.logConsumeFn=logConsume
	}else{
		fmt.Println("===========Ignore Initiation Processed Item Log======")
	}
	//initial cache(redis)
	if !sub.no_cache{
		fmt.Println("===========Initiation Redis for delete Uid: True======")
		var errc *e.Error
		sub.Redis,errc=redis.NewCacheHelper(vault)
		if errc!=nil{
			return errc
		}
	}else{
		fmt.Println("===========Initiation Redis for delete Uid: False======")
	}
	fmt.Println("=>No inject: ",sub.no_inject)
	if sub.logConsumeFn==nil{
		fmt.Println("=>Log consumedFn: True")
	}else{
		fmt.Println("=>Log consumedFn: False")
	}
	
	if sub.RetryDeleteRedisFn==nil{
		fmt.Println("=>RetryDeleteRedisFn: False")
	}else{
		fmt.Println("=>RetryDeleteRedisFn: True")
	}
	return nil
}
//consume message
func  (sub *Subscriber)Consume() (*e.Error){
	//healthSRV:=health.NewHealth("8080")
	log.Info(fmt.Sprintf("Number of cosumer: %s",utils.ItoString(sub.num_consumer)))
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//wg := &sync.WaitGroup{}
	//
	//conf:=NewConfig(config_map)
	//fmt.Printf("%+v",config_map)
	brokers_str:=sub.config["BROKERS"]
	brokers:=utils.Explode(brokers_str,",")
	consumer_group:=sub.config["CONSUMER_GROUP"]
	
	for i:=0; i<sub.num_consumer; i++{
		//config
		conf:=NewConsumerConfig(sub.config)
		conf.Consumer.Offsets.Initial = sarama.OffsetOldest
		//sarama.OffsetOldest get from last offset not yet commit
		//sarama.OffsetNewest  ignore all mesage just get new message after consumer start
		//var err Error
		sub.subscriber, _= kafka.NewSubscriber( //reconfig subscriber because want to each goroutine(subscriber has differerence client_ID)
			kafka.SubscriberConfig{
				Brokers:                brokers,
				Unmarshaler:           kafka.DefaultMarshaler{},
				OverwriteSaramaConfig: conf,
				ConsumerGroup:        consumer_group,
				ReconnectRetrySleep: 10 * time.Second,
			},
			//watermill.NewStdLogger(false, false),
			nil,
		)
		/*if err!=nil{
			log.Error(err.Error(),"ConsumeData","NewConsumerGoroutine",sub.config)
			return nil
		}*/
		//
		messages, err:= sub.subscriber.Subscribe(context.Background(), sub.topic)
		if err!=nil{
			return e.New(err.Error(),"KAFKA","CONSUMER") 
		}
		//
		//wg.Add(1)
		go sub.ProcessMesasge(i+1,messages)
	}
	//health server check
	/*go func(){
		err := healthSRV.ListenAndServe()
		if err != nil {
			log.Error("Stop Health Server","Consume")
		}	
	}()*/
	//
	c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//
	close(c)
	//wg.Wait()          // Block here until are workers are done
	//
	sub.subscriber.Close() //gracefull shutdown, wait all subcriber done task. 
	//
	/*ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := healthSRV.Shutdown(ctx); err != nil {
        // handle err
    }*/
	//
	log.Info("Consumer stop","Consumer")
	return nil
}

func (sub *Subscriber)ProcessMesasge(i int,messages <-chan *message.Message){
	//defer *health=false
	log.Info(fmt.Sprintf("Consumer: %s-[%s] started",sub.id,utils.ItoString(i)))
	
	for msg := range messages {
		
		//process message
		if !sub.no_inject{
			err:=InjectComsumeTime(msg)
			if err!=nil{
				log.Error(err.Msg(),"Consumer","ProcessMesasge")
				//return
			}
		}
		event,err:=ExtractEvent(msg)
		//check ignore reprocess
		
		event.ProcessedFlow=event.ProcessedFlow+"->"+sub.id
		event.WorkerID=i
		if err!=nil{
			log.Error(err.Msg(),"Consumer","ProcessMesasge",event)
		}
		//
		//delete UID from cache server(Redis) after consume and
		//before process for sure if any error when process message, still can delete from redis
		//
		if event.Uid!="" && !sub.no_cache && sub.Redis!=nil{
			res, err:= sub.Redis.Exists(event.Uid)
			if err!=nil{
				event.IgnoreUid=true
				log.Warn(err.Msg(),"ConsumeMessage","CheckRedisExistKey",event)	
				//=> push to kafka, wait for worker recheck & delete
				if sub.RetryDeleteRedisFn!=nil{
					sub.RetryDeleteRedisFn(event)
				}
				//sleep wait for worker delete redis first??? => item still have time for process so temp do not do any more

			}else if res{//exist key
				log.Info("Exists Redis Key: "+event.Uid,"ConsumeMessage","CheckRedisExistKeyBeforeDelete",event)
				err:=sub.DeleteUniqueKey(event)
				if err!=nil{
					//set kafka item fail
					event.IgnoreUid=true
					//=>push to Kafka, wait for worker delete
					if sub.RetryDeleteRedisFn!=nil{
						sub.RetryDeleteRedisFn(event)
					}
					//sleep wait for worker delete redis first??? => item still have time for process so temp do not do any more

					//
					log.Warn(err.Msg(),"ConsumeMessage","RedisDeleteUniqueKey",event)					
				}else{
					log.Info("Delete redis key is success","ConsumeMessage","RedisDeleteUniqueKey",event)
				}
			}else{
				log.Warn("Redis key not exist:"+event.Uid,"ConsumeMessage","RedisDeleteUniqueKeyNotExisit",event)	
			}
		}
		// ignore message if reprocess

		//process message 
		consumed_log:=true
		err_p:=sub.ProcessFn(msg)// => callback Fn
		//if process finished without no error (mean ACK was send)
		//-set finish time
		if err_p!=nil{
			// Log error
			consumed_log=false
			log.Error(err_p.Error(),"Consumer","ProcessMesasge",event)
			event.Logs=event.Logs+"\r\n"+err_p.Error()
		}
		//
		
		//
		
		if sub.no_ack==false{//alway send ack in PRD/STG
			msg.Ack()
			//consumed_log=true
			if err_p!=nil{//Repush Event to main bus if process function has error
				if sub.RePushEventFn!=nil{
					err:=sub.RePushEventFn(event)
					if err!=nil{
						log.Error(err.Msg(),"Consumer","ProcessMesasge",event)
					}
				}
			}
		}else if err_p==nil{//local base on error, if error ==nil send ack 
			msg.Ack()
			//consumed_log=true
		}
		//-push item processed to kafka_item_succes, for kafka_item_fail push from router when TTL
		if !sub.no_inject{
			InjectFinishTime(&event)
			if sub.logConsumeFn!=nil && consumed_log{
				err:=sub.logConsumeFn(event)
				if err!=nil{
					log.Error(err.Error(),"Consumer","ProcessMesasge")
					//return
				}
			}
		}
		
		
	}
	log.Info(fmt.Sprintf("Consumer: %s-[%s] shutdown",sub.id,utils.ItoString(i)))
}
func (sub *Subscriber)SetNoAck(no_ack bool){
	sub.no_ack=no_ack
}

func (sub *Subscriber)SetNoInject(no_inject bool){
	sub.no_inject=no_inject
}
func (sub *Subscriber)SetNoCache(no_cache bool){
	sub.no_cache=no_cache
}
func (sub *Subscriber)SetPushlisher(fn event.RePushFn){
	sub.RePushEventFn=fn
}
func (sub *Subscriber)SetPushlisherForRetryDeleteRedis(fn event.RetryDeleteRedisPushFn){
	sub.RetryDeleteRedisFn=fn
}
func (sub *Subscriber) DeleteUniqueKey(event ev.Event)*e.Error{
	if event.Uid!=""{
		err:=sub.Redis.Del(event.Uid)
		if err!=nil{
			return err 
		}
	}
	return nil
}
func (sub *Subscriber) Clean(){
	sub.Redis.Close()
}
func ExtractEvent(messages *message.Message) (ev.Event,*e.Error){
	//
	event:=ev.Event{}
	err := json.Unmarshal([]byte(messages.Payload), &event)
	if err != nil {
		return event,e.New(err.Error(),"KAFKA","EXTRACT_EVENT") 
	}
	return event,nil
	//
}
func InjectComsumeTime(messages *message.Message) *e.Error{
	//
	event,err:=ExtractEvent(messages)
	if err!=nil{
		return err
	}
	event.ConsumeTime=time.Now()
	//
	data, err_m := json.Marshal(&event)
    if err_m != nil {
        //log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return e.New(err_m.Error(),"EVENT_DRIVEN","MARSHAL_EVENT")
    }
	messages.Payload=[]byte(data)
	return nil
	//msg := message.NewMessage(watermill.NewUUID(), data)
}
func InjectWorkerName(messages *message.Message,worker_name string) *e.Error{
	//
	event,err:=ExtractEvent(messages)
	if err!=nil{
		return err
	}
	event.ProcessedFlow=event.ProcessedFlow+"->"+worker_name
	//
	data, err_m := json.Marshal(&event)
    if err_m != nil {
        //log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return e.New(err_m.Error(),"EVENT_DRIVEN","MARSHAL_EVENT")
    }
	messages.Payload=[]byte(data)
	return nil
	//msg := message.NewMessage(watermill.NewUUID(), data)
}
func InjectFinishTime(event *ev.Event){
	if event!=nil{
		event.FinishTime=time.Now()
		event.ProcessingTime=(time.Now().Sub(event.ConsumeTime).Seconds())
	}
}
