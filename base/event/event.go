package event
import (
	"time"
	"github.com/google/uuid"
	"github.com/ThreeDotsLabs/watermill/message"
	e "github.com/goonma/sdk/base/error"
)
type ConsumeFn = func(message *message.Message) error
type WriteLogConsumeFn = func(e Event) error
type RePushFn =func(event Event) *e.Error
type RetryDeleteRedisPushFn =func(event Event) error
type Event struct {
	EventID     uuid.UUID
	EventName   string//dev set
	EventData   interface{}//dev set
	Uid         string //dev set
	SourceID    string
	WorkerID    int
	Flow	*Queue
	Counter int
	PushlishTime time.Time
	ConsumeTime time.Time
	FinishTime time.Time
	ProcessingTime float64
	Logs string
	ProcessedFlow string
	IgnoreUid bool
	Transaction_id uuid.UUID
	Transaction_start_time int64 //timestamp

}
type Queue []string

func (self *Queue) Push(x string) {
    *self = append(*self, x)
}

func (self *Queue) Pop() string {
    h := *self
    var el string
    l := len(h)
    el, *self = h[0], h[1:l]
    return el
}

func NewQueue() *Queue {
    return &Queue{}
}