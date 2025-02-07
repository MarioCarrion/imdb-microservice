//go:build kafka

package main

import (
	cmdinternal "github.com/MarioCarrion/todo-api-microservice-example/cmd/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/kafka"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/service"
)

// KafkaMessageBroker represents Kafka as a Message Broker.
type KafkaMessageBroker struct {
	producer  *cmdinternal.KafkaProducer
	publisher service.TaskMessageBrokerPublisher
}

// NewMessageBrokerPublisher initializes a new Kafka Broker.
func NewMessageBrokerPublisher(conf *envvar.Configuration) (MessageBrokerPublisher, error) { //nolint: ireturn
	producer, err := cmdinternal.NewKafkaProducer(conf)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "internal.NewKafkaProducer")
	}

	return &KafkaMessageBroker{
		producer:  producer,
		publisher: kafka.NewTask(producer.Producer, producer.Topic),
	}, nil
}

// Publisher returns the Kafka broker.
func (m *KafkaMessageBroker) Publisher() service.TaskMessageBrokerPublisher { //nolint: ireturn
	return m.publisher
}

// Close closes the broker.
func (m *KafkaMessageBroker) Close() error {
	m.producer.Producer.Close()

	return nil
}
