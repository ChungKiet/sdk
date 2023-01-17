package eventdriven

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/goonma/sdk/base/event"
	ev "github.com/goonma/sdk/base/event"
	r "github.com/goonma/sdk/cache/redis"
	"github.com/goonma/sdk/config/vault"
	//"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/pubsub/kafka"

	//"github.com/ThreeDotsLabs/watermill/message"
	"time"

	e "github.com/goonma/sdk/base/error"
	//"fmt"
)

// unit receive from BUS(kafka pubsub)
type EventDriven struct {
	id                   string
	publisher            kafka.Publisher
	subscriber           kafka.Subscriber
	redis                r.CacheHelper
	un_set_pushlish_time bool
}

// intial publisher
func (ev *EventDriven) InitialPublisher(vault *vault.Vault, config_path string, service_id string, args ...interface{}) *e.Error {
	ev.id = service_id
	err := ev.publisher.Initial(vault, config_path, args...)
	if err != nil {
		return err
	}
	//
	//
	return nil
}
func (ev *EventDriven) InitialPublisherWithGlobal(vault *vault.Vault, config_path string, service_id string, args ...interface{}) *e.Error {
	ev.id = service_id
	err := ev.publisher.InitialWithGlobal(vault, config_path, args...)
	if err != nil {
		return err
	}
	return nil
}

// initial subscriber
func (ev *EventDriven) InitialSubscriber(vault *vault.Vault, config_path string, service_id string, callbackfn event.ConsumeFn, logConsume event.WriteLogConsumeFn, args ...interface{}) *e.Error {
	ev.id = service_id
	err := ev.subscriber.Initial(vault, config_path, ev.id, callbackfn, logConsume)
	if err != nil {
		return err
	}
	//
	return nil
}

func (ev *EventDriven) InitialSubscriberWithGlobal(vault *vault.Vault, config_path string, service_id string, callbackfn event.ConsumeFn, logConsume event.WriteLogConsumeFn, replaceTopic string, args ...interface{}) *e.Error {
	ev.id = service_id
	err := ev.subscriber.InitialWithGlobal(vault, config_path, ev.id, callbackfn, logConsume, replaceTopic)
	if err != nil {
		return err
	}
	//
	return nil
}

func (ev *EventDriven) SetNoEvent(v bool) {
	ev.subscriber.SetNoInject(v)
}
func (ev *EventDriven) SetNoUpdatePublishTime(v bool) {
	ev.un_set_pushlish_time = v
}

func (ev *EventDriven) SetPublisherForSubscriber() {
	ev.subscriber.SetPushlisher(ev.Publish)
}

// publish event to BUS
func (ev *EventDriven) Publish(event ev.Event) *e.Error {
	event.EventID = uuid.New()
	if !ev.un_set_pushlish_time {
		event.PushlishTime = time.Now()
	}
	event.ProcessedFlow = event.ProcessedFlow + "->" + ev.id
	if event.Transaction_start_time == 0 { //create new txn
		event.Transaction_id = uuid.New()
		event.Transaction_start_time = time.Now().Unix()
	}
	//serialize event
	data, err := json.Marshal(event)
	if err != nil {
		//log.Error(err.Error(),"EVENT_DRIVEN_SERIALIZE")
		return e.New(err.Error(), "EVENT_DRIVEN", "PUBLISH")
	}
	return ev.publisher.Publish(data)
}

// get event from BUS
func (ev *EventDriven) Subscribe() *e.Error {
	err := ev.subscriber.Consume()
	if err != nil {
		return err
	}
	return nil
}
func (ev *EventDriven) Clean() {
	//if ev.subscriber!=nil{
	ev.subscriber.Clean()
	//}
}

/*
Process Event Template
func ProcessFn(messages <-chan *message.Message) {
	for msg := range messages {
		log.Printf("received message: %s, payload: %s", msg.UUID, string(msg.Payload))

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
*/
