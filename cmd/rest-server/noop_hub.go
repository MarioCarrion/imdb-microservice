//go:build !(redis || rabbitmq || kafka)

package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/envvar"
	"github.com/MarioCarrion/todo-api/internal/service"
)

type NoOpMessageBus struct {
	repo service.TaskMessageBrokerRepository
}

func NewMessageHub(_ *envvar.Configuration, logger *zap.Logger) (MessageBus, error) {
	return &NoOpMessageBus{
		repo: &noOpRepository{logger},
	}, nil
}

func (m *NoOpMessageBus) Repository() service.TaskMessageBrokerRepository {
	return m.repo
}

func (m *NoOpMessageBus) Close() error {
	return nil
}

type noOpRepository struct {
	logger *zap.Logger
}

func (n *noOpRepository) Created(_ context.Context, task internal.Task) error {
	n.logger.Info("noop created", zap.String("task", fmt.Sprintf("%+v", task)))

	return nil
}

func (n *noOpRepository) Deleted(_ context.Context, id string) error {
	n.logger.Info("noop deleted", zap.String("task", id))

	return nil
}

func (n *noOpRepository) Updated(_ context.Context, task internal.Task) error {
	n.logger.Info("noop updated", zap.String("task", fmt.Sprintf("%+v", task)))

	return nil
}
