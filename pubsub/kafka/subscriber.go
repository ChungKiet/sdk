package kafka

import (
	"strings"

	"github.com/goonma/sdk/utils"
	//"github.com/goonma/sdk/utils/transform"
	//"github.com/goonma/sdk/health"
	"os"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/base/event"
	ev "github.com/goonma/sdk/base/event"
	"github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/log"

	//"errors"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/signal"
	"syscall"
	"time"
	//"errors"
)

// type ConsumeFn = func(messages <-chan *message.Message)
type Subscriber struct {
	id            string
	subscriber    *kafka.Subscriber
	topic         string
	ProcessFn     event.ConsumeFn
	RePushEventFn event.RePushFn
	Redis         redis.CacheHelper
	logConsumeFn  event.WriteLogConsumeFn
	num_consumer  int
	no_ack        bool
	no_inject     bool
	//
	config map[string]string
	init_mode int
}

// Initial Publisher
func (sub *Subscriber) Initial(vault *vault.Vault, config_path string, worker_name string, callbackfn event.ConsumeFn, logConsume event.WriteLogConsumeFn) *e.Error {
	log.Info("Initialing Kafka subscriber...", "KAFKA")
	//
	hostname, err_h := os.Hostname()
	if err_h != nil {
		log.Warn(fmt.Sprintf("Error get Hostname: %s", err_h.Error()))
	}
	//prd, stg no resend
	if os.Getenv("ENV") == "prd" || os.Getenv("ENV") == "stg" || os.Getenv("ENV") == "dev" {
		sub.SetNoAck(false)
	} else { //local ENv resend message
		sub.SetNoAck(true)
	}
	sub.id = worker_name + "_" + hostname
	//

	//
	config_map := GetConfig(vault, config_path)
	sub.config = config_map
	//conf:=NewConfig(config_map)
	//fmt.Printf("%+v",config_map)
	brokers_str := config_map["BROKERS"]
	topic := config_map["TOPIC"]
	consumer_group := config_map["CONSUMER_GROUP"]
	num_consumer := utils.ItoInt(config_map["NUM_CONSUMER"])
	if num_consumer == math.MinInt32 {
		return e.New("Event Bus Number of Consumer must be number", "KAFKA", "CONSUMER")
	}
	sub.num_consumer = num_consumer

	if brokers_str == "" {
		//log.ErrorF("Event Bus Brokers not found","KAFKA","KAFKA_CONSUMER_BROKER")
		return e.New("Event Bus Brokers not found", "KAFKA", "CONSUMER")
	}
	if topic == "" {
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_CONSUMER_TOPIC")
		return e.New("Event Bus Topic not found", "KAFKA", "CONSUMER")
	}
	if consumer_group == "" {
		//log.ErrorF("Event Bus Consumer Group not found","KAFKA","CONSUMER")
		return e.New("Event Bus Consumer Group not found", "KAFKA", "CONSUMER")
	}

	sub.topic = topic
	//
	var err error
	brokers := utils.Explode(brokers_str, ",")
	//conf:=kafka.DefaultSaramaSubscriberConfig()
	conf := NewConsumerConfig(config_map)
	conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	//sarama.OffsetOldest get from last offset not yet commit
	//sarama.OffsetNewest  ignore all mesage just get new message after consumer start
	sub.subscriber, err = kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
			ConsumerGroup:         consumer_group,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)
	if err != nil {
		if log.LogMode() != 0 {
			data_cfg_str, err_c := utils.MapToJSONString(config_map)
			if err_c != nil {
				//log.ErrorF(err.Error(),"KAFKA")
				return e.New(err_c.Error()+": "+data_cfg_str, "KAFKA", "CONSUMER")
			}
			return e.New(fmt.Sprintf("%s: %s", err.Error(), data_cfg_str), "KAFKA", "CONSUMER")
		} else {
			return e.New(err.Error(), "KAFKA", "CONSUMER")
		}
	}
	log.Info(fmt.Sprintf("%s %s: %s", "Kafka consumer brokers: ", brokers_str+" | "+topic, " connected"), "KAFKA", "CONSUMER")
	sub.ProcessFn = callbackfn
	if logConsume != nil {
		fmt.Println("===========Initiation Processed Item Log======")
		sub.logConsumeFn = logConsume
	} else {
		fmt.Println("===========Ignore Initiation Processed Item Log======")
	}
	//	fmt.Println("===========Initiation Redis for delete Uid: True======")
	fmt.Println("===========Initiation Redis for check other pod ready======")
	var errc *e.Error
	sub.Redis, errc = redis.NewCacheHelper(vault)

	if errc != nil {
		return errc
	}
	fmt.Println("=>No inject: ", sub.no_inject)
	if sub.logConsumeFn == nil {
		fmt.Println("=>Log consumedFn: True")
	} else {
		fmt.Println("=>Log consumedFn: False")
	}
	sub.init_mode=0
	return nil
}

func (sub *Subscriber) InitialWithGlobal(vault *vault.Vault, config_path string, worker_name string, callbackfn event.ConsumeFn, logConsume event.WriteLogConsumeFn, replaceTopic string) *e.Error {
	log.Info("Initialing Kafka subscriber...", "KAFKA")
	if config_path == "" {
		return e.New("Vault config path for Kafka is empty", "KAFKA", "PUBLISHER")
	}
	hostname, err_h := os.Hostname()
	if err_h != nil {
		log.Warn(fmt.Sprintf("Error get Hostname: %s", err_h.Error()))
	}
	//prd, stg no resend
	if os.Getenv("ENV") == "prd" || os.Getenv("ENV") == "stg" || os.Getenv("ENV") == "dev" {
		sub.SetNoAck(false)
	} else { //local ENv resend message
		sub.SetNoAck(true)
	}
	sub.id = worker_name + "_" + hostname
	arr := utils.Explode(config_path, "/")
	global_config_path := strings.Join(arr[:len(arr)-1], "/")
	global_config_map := GetConfig(vault, fmt.Sprintf("%s/%s", global_config_path, "general"))

	var config_map map[string]string
	check, _ := vault.CheckItemExist(config_path)
	if check {
		local_config_map := GetConfig(vault, config_path)
		config_map = MergeConfig(global_config_map, local_config_map)
	} else {
		config_map = global_config_map
	}
	consumer_group := fmt.Sprintf("%s-%s-consumer-group", hostname, config_map["TOPIC"])
	config_map["CONSUMER_GROUP"] = consumer_group
	sub.config = config_map
	brokers_str := config_map["BROKERS"]
	topic := ""
	if replaceTopic == "" {
		topic = config_map["TOPIC"]
	} else {
		topic = replaceTopic
	}
	num_consumer := utils.ItoInt(config_map["NUM_CONSUMER"])
	if num_consumer == math.MinInt32 {
		return e.New("Event Bus Number of Consumer must be number", "KAFKA", "CONSUMER")
	}
	sub.num_consumer = num_consumer

	if brokers_str == "" {
		//log.ErrorF("Event Bus Brokers not found","KAFKA","KAFKA_CONSUMER_BROKER")
		return e.New("Event Bus Brokers not found", "KAFKA", "CONSUMER")
	}
	if topic == "" {
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_CONSUMER_TOPIC")
		return e.New("Event Bus Topic not found", "KAFKA", "CONSUMER")
	}
	sub.topic = topic
	var err error
	brokers := utils.Explode(brokers_str, ",")
	conf := NewConsumerConfig(config_map)
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest
	//sarama.OffsetOldest get from last offset not yet commit
	//sarama.OffsetNewest  ignore all mesage just get new message after consumer start
	sub.subscriber, err = kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
			ConsumerGroup:         consumer_group,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)
	if err != nil {
		if log.LogMode() != 0 {
			data_cfg_str, err_c := utils.MapToJSONString(config_map)
			if err_c != nil {
				//log.ErrorF(err.Error(),"KAFKA")
				return e.New(err_c.Error()+": "+data_cfg_str, "KAFKA", "CONSUMER")
			}
			return e.New(fmt.Sprintf("%s: %s", err.Error(), data_cfg_str), "KAFKA", "CONSUMER")
		} else {
			return e.New(err.Error(), "KAFKA", "CONSUMER")
		}
	}
	log.Info(fmt.Sprintf("%s %s: %s", "Kafka consumer brokers: ", brokers_str+" | "+topic, " connected"), "KAFKA", "CONSUMER")
	sub.ProcessFn = callbackfn
	if logConsume != nil {
		fmt.Println("===========Initiation Processed Item Log======")
		sub.logConsumeFn = logConsume
	} else {
		fmt.Println("===========Ignore Initiation Processed Item Log======")
	}
	var errc *e.Error
	sub.Redis, errc = redis.NewCacheHelper(vault)
	if errc != nil {
		return errc
	}
	fmt.Println("=>No inject: ", sub.no_inject)
	if sub.logConsumeFn == nil {
		fmt.Println("=>Log consumedFn: True")
	} else {
		fmt.Println("=>Log consumedFn: False")
	}
	sub.init_mode=1
	return nil
}

// consume message
func (sub *Subscriber) Consume() *e.Error {
	//healthSRV:=health.NewHealth("8080")
	log.Info(fmt.Sprintf("Number of cosumer: %s", utils.ItoString(sub.num_consumer)))
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//wg := &sync.WaitGroup{}
	//
	//conf:=NewConfig(config_map)
	//fmt.Printf("%+v",config_map)
	brokers_str := sub.config["BROKERS"]
	brokers := utils.Explode(brokers_str, ",")
	consumer_group := sub.config["CONSUMER_GROUP"]
	//wait for other pod
	str_num_pod := sub.config["NUM_POD"]
	if str_num_pod != "" {
		//pod start success => increase number of pod
		_, err := sub.Redis.IncreaseInt(consumer_group, 1)
		if err != nil {
			_, err := sub.Redis.IncreaseInt(consumer_group, 1)
			if err != nil {
				return err
			}
		}
		num_pod := utils.StringToInt(str_num_pod)
		if num_pod > 0 {
			i_current_num_pod, err := sub.Redis.Get(consumer_group)
			if err != nil {
				i_current_num_pod = "0"
			}
			current_num_pod := utils.ItoInt(i_current_num_pod)
			for current_num_pod < num_pod {
				time.Sleep(3 * time.Second)
				fmt.Println("Wait for other pod: ", current_num_pod, "/", num_pod)
				i_current_num_pod, _ = sub.Redis.Get(consumer_group)
				current_num_pod = utils.ItoInt(i_current_num_pod)
			}
		}

	}

	//all pod ready
	for i := 0; i < sub.num_consumer; i++ {
		//config
		conf := NewConsumerConfig(sub.config)
		if sub.init_mode==0{
			conf.Consumer.Offsets.Initial = sarama.OffsetOldest
		}else{
			conf.Consumer.Offsets.Initial = sarama.OffsetNewest
		}
		//sarama.OffsetOldest get from last offset not yet commit
		//sarama.OffsetNewest  ignore all mesage just get new message after consumer start
		//var err Error
		sub.subscriber, _ = kafka.NewSubscriber( //reconfig subscriber because want to each goroutine(subscriber has differerence client_ID)
			kafka.SubscriberConfig{
				Brokers:               brokers,
				Unmarshaler:           kafka.DefaultMarshaler{},
				OverwriteSaramaConfig: conf,
				ConsumerGroup:         consumer_group,
				ReconnectRetrySleep:   10 * time.Second,
			},
			//watermill.NewStdLogger(false, false),
			nil,
		)
		/*if err!=nil{
			log.Error(err.Error(),"ConsumeData","NewConsumerGoroutine",sub.config)
			return nil
		}*/
		//
		messages, err := sub.subscriber.Subscribe(context.Background(), sub.topic)
		if err != nil {
			return e.New(err.Error(), "KAFKA", "CONSUMER")
		}
		//
		//wg.Add(1)
		go sub.ProcessMesasge(i+1, messages)
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
	log.Info("Consumer stop", "Consumer")
	return nil
}

func (sub *Subscriber) ProcessMesasge(i int, messages <-chan *message.Message) {
	//defer *health=false
	log.Info(fmt.Sprintf("Consumer: %s-[%s] started", sub.id, utils.ItoString(i)))

	for msg := range messages {

		//process message
		if !sub.no_inject {
			err := InjectComsumeTime(msg)
			if err != nil {
				log.Error(err.Msg(), "Consumer", "ProcessMesasge")
				//return
			}
		}
		event, err := ExtractEvent(msg)
		//check ignore reprocess

		event.ProcessedFlow = event.ProcessedFlow + "->" + sub.id
		event.WorkerID = i
		if err != nil {
			log.Error(err.Msg(), "Consumer", "ProcessMesasge", event)
		}

		//process message
		consumed_log := true
		err_p := sub.ProcessFn(msg) // => callback Fn
		//if process finished without no error (mean ACK was send)
		//-set finish time
		if err_p != nil {
			// Log error
			consumed_log = false
			log.Error(err_p.Error(), "Consumer", "ProcessMesasge", event)
			event.Logs = event.Logs + "\r\n" + err_p.Error()
		}
		//

		//

		if sub.no_ack == false { //alway send ack in PRD/STG
			msg.Ack()
			//consumed_log=true
			if err_p != nil { //Repush Event to main bus if process function has error
				if sub.RePushEventFn != nil {
					/*err := sub.RePushEventFn(event)
					if err != nil {
						log.Error(err.Msg(), "Consumer", "ProcessMesasge", event)
					}*/
				}
			}
		} else if err_p == nil { //local base on error, if error ==nil send ack
			msg.Ack()
			//consumed_log=true
		}
		//-push item processed to kafka_item_succes, for kafka_item_fail push from router when TTL
		if !sub.no_inject {
			InjectFinishTime(&event)
			if sub.logConsumeFn != nil && consumed_log {
				err := sub.logConsumeFn(event)
				if err != nil {
					log.Error(err.Error(), "Consumer", "ProcessMesasge")
					//return
				}
			}
		}

	}
	log.Info(fmt.Sprintf("Consumer: %s-[%s] shutdown", sub.id, utils.ItoString(i)))
}
func (sub *Subscriber) SetNoAck(no_ack bool) {
	sub.no_ack = no_ack
}

func (sub *Subscriber) SetNoInject(no_inject bool) {
	sub.no_inject = no_inject
}

func (sub *Subscriber) SetPushlisher(fn event.RePushFn) {
	sub.RePushEventFn = fn
}
func (sub *Subscriber) Clean() {
	sub.Redis.Close()
}
func ExtractEvent(messages *message.Message) (ev.Event, *e.Error) {
	//
	event := ev.Event{}
	err := json.Unmarshal([]byte(messages.Payload), &event)
	if err != nil {
		return event, e.New(err.Error(), "KAFKA", "EXTRACT_EVENT")
	}
	return event, nil
	//
}
func InjectComsumeTime(messages *message.Message) *e.Error {
	//
	event, err := ExtractEvent(messages)
	if err != nil {
		return err
	}
	event.ConsumeTime = time.Now()
	//
	data, err_m := json.Marshal(&event)
	if err_m != nil {
		//log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return e.New(err_m.Error(), "EVENT_DRIVEN", "MARSHAL_EVENT")
	}
	messages.Payload = []byte(data)
	return nil
	//msg := message.NewMessage(watermill.NewUUID(), data)
}
func InjectWorkerName(messages *message.Message, worker_name string) *e.Error {
	//
	event, err := ExtractEvent(messages)
	if err != nil {
		return err
	}
	event.ProcessedFlow = event.ProcessedFlow + "->" + worker_name
	//
	data, err_m := json.Marshal(&event)
	if err_m != nil {
		//log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return e.New(err_m.Error(), "EVENT_DRIVEN", "MARSHAL_EVENT")
	}
	messages.Payload = []byte(data)
	return nil
	//msg := message.NewMessage(watermill.NewUUID(), data)
}
func InjectFinishTime(event *ev.Event) {
	if event != nil {
		event.FinishTime = time.Now()
		event.ProcessingTime = (time.Now().Sub(event.ConsumeTime).Seconds())
	}
}
