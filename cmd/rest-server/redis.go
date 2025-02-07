//go:build !kafka && !rabbitmq

package main

import (
	"github.com/go-redis/redis/v8"

	cmdinternal "github.com/MarioCarrion/todo-api-microservice-example/cmd/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
	internalredis "github.com/MarioCarrion/todo-api-microservice-example/internal/redis"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/service"
)

// RedisMessageBroker represents Redis as a Message Broker.
type RedisMessageBroker struct {
	client    *redis.Client
	publisher service.TaskMessageBrokerPublisher
}

// NewMessageBrokerPublisher initializes a new Redis Broker.
func NewMessageBrokerPublisher(conf *envvar.Configuration) (MessageBrokerPublisher, error) { //nolint: ireturn
	producer, err := cmdinternal.NewRedis(conf)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "internal.NewRedis")
	}

	return &RedisMessageBroker{
		client:    producer,
		publisher: internalredis.NewTask(producer),
	}, nil
}

// Publisher returns the Redis broker.
func (m *RedisMessageBroker) Publisher() service.TaskMessageBrokerPublisher { //nolint: ireturn
	return m.publisher
}

// Close closes the broker.
func (m *RedisMessageBroker) Close() error {
	if err := m.client.Close(); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "producer.Close")
	}

	return nil
}
