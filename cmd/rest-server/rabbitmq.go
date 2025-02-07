//go:build rabbitmq

package main

import (
	cmdinternal "github.com/MarioCarrion/todo-api-microservice-example/cmd/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/rabbitmq"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/service"
)

// RabbitMQMessageBroker represents RabbitMQ as a Message Broker.
type RabbitMQMessageBroker struct {
	producer  *cmdinternal.RabbitMQ
	publisher service.TaskMessageBrokerPublisher
}

// NewMessageBrokerPublisher initializes a new RabbitMQ Broker.
func NewMessageBrokerPublisher(conf *envvar.Configuration) (MessageBrokerPublisher, error) { //nolint: ireturn
	client, err := cmdinternal.NewRabbitMQ(conf)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "internal.NewRabbitMQ")
	}

	return &RabbitMQMessageBroker{
		producer:  client,
		publisher: rabbitmq.NewTask(client.Channel),
	}, nil
}

// Publisher returns the RabbitMQ broker.
func (m *RabbitMQMessageBroker) Publisher() service.TaskMessageBrokerPublisher { //nolint: ireturn
	return m.publisher
}

// Close closes the broker.
func (m *RabbitMQMessageBroker) Close() error {
	if err := m.producer.Close(); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "producer.Close")
	}

	return nil
}
