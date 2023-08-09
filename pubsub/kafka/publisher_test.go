package kafka

import (
	"reflect"
	"testing"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	e "github.com/goonma/sdk/base/error"
)

func TestPublisher_PublishWithPartition(t *testing.T) {
	type fields struct {
		publisherWithPartitioning *kafka.Publisher
		topic                     string
	}
	type args struct {
		data []byte
		key  string
	}

	marshaler := kafka.NewWithPartitioningMarshaler(func(topic string, msg *message.Message) (string, error) {
		return msg.Metadata.Get("message_key"), nil
	})

	publisherWithPartitioning, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   []string{"localhost:29092"},
			Marshaler: marshaler,
		},
		//watermill.NewStdLogger(false, false),
		nil,
	)

	if err != nil {
		t.Errorf("Publisher.PublishWithPartition() = %v, want %v", err, nil)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *e.Error
	}{
		{
			name: "send message with key test1",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test1",
			},
			want: nil,
		},
		{
			name: "send message with key test2",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test2",
			},
			want: nil,
		},
		{
			name: "send message with key test3",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test3",
			},
			want: nil,
		},
		{
			name: "send message with key test1 again",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test1",
			},
			want: nil,
		},
		{
			name: "send message with key test2 again",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test2",
			},
			want: nil,
		},
		{
			name: "send message with key test3 again",
			fields: fields{
				publisherWithPartitioning: publisherWithPartitioning,
				topic:                     "phucnh.test",
			},
			args: args{
				data: []byte("test"),
				key:  "test3",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub := &Publisher{
				publisherWithPartitioning: tt.fields.publisherWithPartitioning,
				topic:                     tt.fields.topic,
			}
			if got := pub.PublishWithPartitioning(tt.args.data, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Publisher.PublishWithPartition() = %v, want %v", got, tt.want)
			}
		})
	}
}
