package kafka

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/utils"

	//"errors"
	"fmt"
	"strings"
)

type Publisher struct {
	publisher              *kafka.Publisher
	publisherWithPartitioning *kafka.Publisher
	topic                  string
}

// Initial Publisher
func (pub *Publisher) Initial(vault *vault.Vault, config_path string, args ...interface{}) *e.Error {
	log.Info("Initialing Kafka publisher...", "KAFKA")
	if config_path == "" {
		return e.New("Vault config path for Kafka is empty", "KAFKA", "PUBLISHER")
	}
	//global event bus config path
	global_config_map := GetConfig(vault, config_path)
	//get service name
	service_name := vault.GetServiceName()
	service_config_path := strings.ReplaceAll(service_name, ".", "/")
	//local event bus config path
	evenbus_config_path := fmt.Sprintf("%s/%s", service_config_path, "eventbus")
	config_map := utils.DictionaryString()
	check_path, err_c := vault.CheckPathExist(evenbus_config_path)
	if err_c != nil {
		log.Error(err_c.Msg(), err_c.Group(), err_c.Key())
	}
	if check_path {
		local_config_map := GetConfig(vault, evenbus_config_path)
		config_map = MergeConfig(global_config_map, local_config_map)
	} else {
		config_map = global_config_map
	}
	//
	conf := NewProducerConfig(config_map)
	brokers_str := config_map["BROKERS"]
	topic := config_map["TOPIC"]
	if brokers_str == "" {
		//log.ErrorF("Event Bus Brokers not found","KAFKAL","KAFKA_PUBLISHER_BROKER")
		return e.New("Event Bus Brokers not found", "KAFKA", "PUBLISHER")
	}
	if topic == "" {
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_PUBLISHER_TOPIC")
		return e.New("Event Bus Topic not found", "KAFKA", "PUBLISHER")
	}
	pub.topic = topic
	var err error

	brokers := utils.Explode(brokers_str, ",")
	pub.publisher, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)

	if err != nil {
		return e.New(err.Error(), "KAFKA", "PUBLISHER")
	}

	marshaler := kafka.NewWithPartitioningMarshaler(func(topic string, msg *message.Message) (string, error) {
		return msg.Metadata.Get("message_key"), nil
	})

	pub.publisherWithPartitioning, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             marshaler,
			OverwriteSaramaConfig: conf,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)

	if err != nil {
		return e.New(err.Error(), "KAFKA", "PUBLISHER")
	}

	//not production
	if err != nil {
		if log.LogMode() != 0 {
			data_cfg_str, err_c := utils.MapToJSONString(config_map)
			if err_c != nil {
				//log.ErrorF(err.Error(),"KAFKA_PUBLISHER_INITIAL")
				return e.New(err_c.Error(), "KAFKA", "PUBLISHER")
			}
			//return errors.New(fmt.Sprintf("%s: %s",err.Error(),data_cfg_str))
			return e.New(fmt.Sprintf("%s: %s", err.Error(), data_cfg_str), "KAFKA", "PUBLISHER")
		} else {
			return e.New(err.Error(), "KAFKA", "PUBLISHER")
		}
	}
	publisher_name := ""
	if len(args) > 0 {
		publisher_name = fmt.Sprintf("[%s]", utils.ItoString(args[0]))
	}
	log.Info(fmt.Sprintf("%s %s %s %s: %s", "Kafka publisher ", publisher_name, " brokers: ", brokers_str+" | "+topic, " connected"), "KAFKA", "PUBLISHER")
	return nil
}

// publish message
func (pub *Publisher) Publish(data []byte) *e.Error {
	msg := message.NewMessage(watermill.NewUUID(), data)
	err := pub.publisher.Publish(pub.topic, msg)
	if err != nil {
		return e.New(err.Error(), "KAFKA", "PUBLISHER")
	}
	return nil
}

// publish message with partition
func (pub *Publisher) PublishWithPartitioning(data []byte, key string) *e.Error {
	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("message_key", key)
	err := pub.publisherWithPartitioning.Publish(pub.topic, msg)
	if err != nil {
		return e.New(err.Error(), "KAFKA", "PUBLISHER")
	}
	return nil
}

// Initial Publisher with global config
func (pub *Publisher) InitialWithGlobal(vault *vault.Vault, config_path string, args ...interface{}) *e.Error {
	log.Info("Initialing Kafka publisher...", "KAFKA")
	if config_path == "" {
		return e.New("Vault config path for Kafka is empty", "KAFKA", "PUBLISHER")
	}
	//global event bus config path
	arr := utils.Explode(config_path, "/")
	global_config_path := strings.Join(arr[:len(arr)-1], "/")
	global_config_map := GetConfig(vault, fmt.Sprintf("%s/%s", global_config_path, "general"))

	config_map := utils.DictionaryString()
	check, _ := vault.CheckItemExist(config_path)
	if check {
		local_config_map := GetConfig(vault, config_path)
		config_map = MergeConfig(global_config_map, local_config_map)
	} else {
		config_map = global_config_map
	}
	//
	conf := NewProducerConfig(config_map)
	brokers_str := config_map["BROKERS"]
	topic := config_map["TOPIC"]
	if brokers_str == "" {
		//log.ErrorF("Event Bus Brokers not found","KAFKAL","KAFKA_PUBLISHER_BROKER")
		return e.New("Event Bus Brokers not found", "KAFKA", "PUBLISHER")
	}
	if topic == "" {
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_PUBLISHER_TOPIC")
		return e.New("Event Bus Topic not found", "KAFKA", "PUBLISHER")
	}
	pub.topic = topic
	var err error

	brokers := utils.Explode(brokers_str, ",")
	pub.publisher, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)
	//not production
	if err != nil {
		if log.LogMode() != 0 {
			data_cfg_str, err_c := utils.MapToJSONString(config_map)
			if err_c != nil {
				//log.ErrorF(err.Error(),"KAFKA_PUBLISHER_INITIAL")
				return e.New(err_c.Error(), "KAFKA", "PUBLISHER")
			}
			//return errors.New(fmt.Sprintf("%s: %s",err.Error(),data_cfg_str))
			return e.New(fmt.Sprintf("%s: %s", err.Error(), data_cfg_str), "KAFKA", "PUBLISHER")
		} else {
			return e.New(err.Error(), "KAFKA", "PUBLISHER")
		}
	}
	publisher_name := ""
	if len(args) > 0 {
		publisher_name = fmt.Sprintf("[%s]", utils.ItoString(args[0]))
	}
	log.Info(fmt.Sprintf("%s %s %s %s: %s", "Kafka publisher ", publisher_name, " brokers: ", brokers_str, " connected"), "KAFKA", "PUBLISHER")
	return nil
}

func (pub *Publisher) InitialManual(config_map map[string]string, publisher_name string) *e.Error {
	log.Info("Initialing Kafka publisher...", "KAFKA")
	//
	conf := NewProducerConfig(config_map)
	brokers_str := config_map["BROKERS"]
	topic := config_map["TOPIC"]
	if brokers_str == "" {
		//log.ErrorF("Event Bus Brokers not found","KAFKAL","KAFKA_PUBLISHER_BROKER")
		return e.New("Event Bus Brokers not found", "KAFKA", "PUBLISHER")
	}
	if topic == "" {
		//log.ErrorF("Event Bus Topic not found","KAFKA","KAFKA_PUBLISHER_TOPIC")
		return e.New("Event Bus Topic not found", "KAFKA", "PUBLISHER")
	}
	pub.topic = topic
	var err error

	brokers := utils.Explode(brokers_str, ",")
	pub.publisher, err = kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: conf,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)
	//not production
	if err != nil {
		if log.LogMode() != 0 {
			data_cfg_str, err_c := utils.MapToJSONString(config_map)
			if err_c != nil {
				//log.ErrorF(err.Error(),"KAFKA_PUBLISHER_INITIAL")
				return e.New(err_c.Error(), "KAFKA", "PUBLISHER")
			}
			//return errors.New(fmt.Sprintf("%s: %s",err.Error(),data_cfg_str))
			return e.New(fmt.Sprintf("%s: %s", err.Error(), data_cfg_str), "KAFKA", "PUBLISHER")
		} else {
			return e.New(err.Error(), "KAFKA", "PUBLISHER")
		}
	}
	log.Info(fmt.Sprintf("%s %s %s %s: %s", "Kafka publisher ", publisher_name, " brokers: ", brokers_str, " connected"), "KAFKA", "PUBLISHER")
	return nil
}
