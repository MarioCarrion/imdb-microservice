//go:build kafka

package main

import (
	"go.uber.org/zap"

	internaldomain "github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/envvar"
	"github.com/MarioCarrion/todo-api/internal/kafka"
	"github.com/MarioCarrion/todo-api/internal/service"

	cmdinternal "github.com/MarioCarrion/todo-api/cmd/internal"
)

type KafkaMessageHub struct {
	producer *cmdinternal.KafkaProducer
	repo     service.TaskMessageBrokerRepository
}

// NewMessageHub initializes a new Kafka Hub.
func NewMessageHub(conf *envvar.Configuration, _ *zap.Logger) (MessageBus, error) {
	producer, err := cmdinternal.NewKafkaProducer(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewKafka")
	}

	return &KafkaMessageHub{
		producer: producer,
		repo:     kafka.NewTask(producer.Producer, producer.Topic),
	}, nil
}

func (m *KafkaMessageHub) Repository() service.TaskMessageBrokerRepository {
	return m.repo
}

func (m *KafkaMessageHub) Close() error {
	m.producer.Producer.Close()
	return nil
}
