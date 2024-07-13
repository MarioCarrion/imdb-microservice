//go:build redis

package main

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	internaldomain "github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/envvar"
	internalredis "github.com/MarioCarrion/todo-api/internal/redis"
	"github.com/MarioCarrion/todo-api/internal/service"

	cmdinternal "github.com/MarioCarrion/todo-api/cmd/internal"
)

type RedisMessageHub struct {
	client *redis.Client
	repo   service.TaskMessageBrokerRepository
}

// NewMessageHub initializes a new Redis Hub.
func NewMessageHub(conf *envvar.Configuration, _ *zap.Logger) (MessageBus, error) {
	rdb, err := cmdinternal.NewRedis(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewRedis")
	}

	return &RedisMessageHub{
		client: rdb,
		repo:   internalredis.NewTask(rdb),
	}, nil
}

func (m *RedisMessageHub) Repository() service.TaskMessageBrokerRepository {
	return m.repo
}

func (m *RedisMessageHub) Close() error {
	return m.client.Close()
}
