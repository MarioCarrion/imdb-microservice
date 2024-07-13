package rabbitmq

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/MarioCarrion/todo-api/internal"
)

const otelName = "github.com/MarioCarrion/todo-api/internal/rabbitmq"

// Task represents the repository used for publishing Task records.
type Task struct {
	ch *amqp.Channel
}

// NewTask instantiates the Task repository.
func NewTask(channel *amqp.Channel) *Task {
	return &Task{
		ch: channel,
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

func (t *Task) publish(ctx context.Context, spanName, routingKey string, event interface{}) error {
	_, span := otel.Tracer(otelName).Start(ctx, spanName)
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   semconv.MessagingSystemKey,
			Value: attribute.StringValue("rabbitmq"),
		},
		attribute.KeyValue{
			Key:   semconv.MessagingRabbitmqRoutingKeyKey,
			Value: attribute.StringValue(routingKey),
		},
	)

	//-

	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(event); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "gob.Encode")
	}

	err := t.ch.Publish(
		"tasks",    // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			AppId:       "tasks-rest-server",
			ContentType: "application/x-encoding-gob", // XXX: We will revisit this in future episodes
			Body:        b.Bytes(),
			Timestamp:   time.Now(),
		})
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ch.Publish")
	}

	return nil
}
