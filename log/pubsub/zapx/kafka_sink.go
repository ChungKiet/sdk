package zapx

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
	"github.com/goonma/sdk/utils"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

var (
	kafkaSinkInsts = map[string]kafkaSink{}
)

type kafkaSink struct {
	kafkaProducer sarama.SyncProducer
	isAsync       bool
	topic         string
}

func getKafkaSink(brokers []string, topic string, config *sarama.Config) kafkaSink {
	producerInst, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	kafkaSinkInst := kafkaSink{
		kafkaProducer: producerInst,
		topic:         topic,
	}
	return kafkaSinkInst
}

// InitKafkaSink  create kafka sink instance
func InitKafkaSink(u *url.URL) (zap.Sink, error) {
	topic := kafkaDefaultTopic
	username:=""
	password:=""
	if t := u.Query().Get("topic"); len(t) > 0 {
		topic = t
	}
	if t := u.Query().Get("username"); len(t) > 0 {
		username = t
	}
	if t := u.Query().Get("password"); len(t) > 0 {
		password = t
	}
	brokers:=utils.Explode(u.Host,",")
	//brokers := []string{arr}
	instKey := strings.Join(brokers, ",")
	if v, ok := kafkaSinkInsts[instKey]; ok {
		return v, nil
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Metadata.Full = true
	//conf.Version = sarama.V0_10_0_0
	//conf.ClientID = "Treehouse-SDK-1.0.1-Publisher-log"
	if username!="" && password!=""{
		config.Net.SASL.Enable = true
		config.Net.SASL.User = username
		config.Net.SASL.Password =password
		config.Net.SASL.Handshake = true
	}
	//
	if ack := u.Query().Get("acks"); len(ack) > 0 {
		if iack, err := strconv.Atoi(ack); err == nil {
			config.Producer.RequiredAcks = sarama.RequiredAcks(iack)
		} else {
			log.Printf("kafka producer acks value '%s' invalid  use default value %d\n", ack, config.Producer.RequiredAcks)
		}
	}
	if retries := u.Query().Get("retries"); len(retries) > 0 {
		if iretries, err := strconv.Atoi(retries); err == nil {
			config.Producer.Retry.Max = iretries
		} else {
			log.Printf("kafka producer retries value '%s' invalid  use default value %d\n", retries, config.Producer.Retry.Max)
		}
	}
	kafkaSinkInsts[instKey] = getKafkaSink(brokers, topic, config)
	return kafkaSinkInsts[instKey], nil
}

// Close implement zap.Sink func Close
func (p kafkaSink) Close() error {
	return nil
}

// Write implement zap.Sink func Write
func (p kafkaSink) Write(b []byte) (n int, err error) {
	var multiErr MultiError
	for _, topic := range strings.Split(p.topic, ",") {
		_, _, err = p.kafkaProducer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(time.Now().String()),
			Value: sarama.ByteEncoder(b),
		})
		if err != nil {
			multiErr = append(multiErr, err)
		}
	}
	return len(b), multiErr
}

// Sync implement zap.Sink func Sync
func (p kafkaSink) Sync() error {
	return nil
}
