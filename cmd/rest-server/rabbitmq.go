//go:build rabbitmq

package main

import (
	"go.uber.org/zap"

	internaldomain "github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/envvar"
	"github.com/MarioCarrion/todo-api/internal/rabbitmq"
	"github.com/MarioCarrion/todo-api/internal/service"

	cmdinternal "github.com/MarioCarrion/todo-api/cmd/internal"
)

type RabbitMQMessageHub struct {
	client *cmdinternal.RabbitMQ
	repo   service.TaskMessageBrokerRepository
}

// NewMessageHub initializes a new RabbitMQ Hub.
func NewMessageHub(conf *envvar.Configuration, _ *zap.Logger) (MessageBus, error) {
	client, err := cmdinternal.NewRabbitMQ(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewRabbitMQ")
	}

	return &RabbitMQMessageHub{
		client: client,
		repo:   rabbitmq.NewTask(client.Channel),
	}, nil
}

func (m *RabbitMQMessageHub) Repository() service.TaskMessageBrokerRepository {
	return m.repo
}

func (m *RabbitMQMessageHub) Close() error {
	return m.client.Close()
}
