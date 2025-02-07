package redis

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

const otelName = "github.com/MarioCarrion/todo-api-microservice-example/internal/redis"

// Task represents the repository used for publishing Task records.
type Task struct {
	client *redis.Client
}

// NewTask instantiates the Task repository.
func NewTask(client *redis.Client) *Task {
	return &Task{
		client: client,
	}
}

// Created publishes a message indicating a task was created.
func (t *Task) Created(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Created", "tasks.event.created", task)
}

// Deleted publishes a message indicating a task was deleted.
func (t *Task) Deleted(ctx context.Context, id string) error {
	return t.publish(ctx, "Task.Deleted", "tasks.event.deleted", id)
}

// Updated publishes a message indicating a task was updated.
func (t *Task) Updated(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Updated", "tasks.event.updated", task)
}

func (t *Task) publish(ctx context.Context, spanName, channel string, event interface{}) error {
	ctx, span := otel.Tracer(otelName).Start(ctx, spanName)
	defer span.End()

	span.SetAttributes(
		semconv.DBSystemRedis,
		attribute.KeyValue{
			Key:   semconv.DBStatementKey,
			Value: attribute.StringValue("PUBLISH"),
		},
	)

	//-

	var b bytes.Buffer

	if err := json.NewEncoder(&b).Encode(event); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.Encode")
	}

	res := t.client.Publish(ctx, channel, b.Bytes())
	if err := res.Err(); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "client.Publish")
	}

	return nil
}
