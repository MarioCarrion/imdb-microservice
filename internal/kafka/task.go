package kafka

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

const otelName = "github.com/MarioCarrion/todo-api-microservice-example/internal/kafka"

// Task represents the Message Broker publisher used to publish Task records.
type Task struct {
	producer  *kafka.Producer
	topicName string
}

type event struct {
	Type  string
	Value internal.Task
}

// NewTask instantiates the Task message broker publisher.
func NewTask(producer *kafka.Producer, topicName string) *Task {
	return &Task{
		topicName: topicName,
		producer:  producer,
	}
}

// Created publishes a message indicating a task was created.
func (t *Task) Created(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Created", "tasks.event.created", task)
}

// Deleted publishes a message indicating a task was deleted.
func (t *Task) Deleted(ctx context.Context, id string) error {
	return t.publish(ctx, "Task.Deleted", "tasks.event.deleted", internal.Task{ID: id})
}

// Updated publishes a message indicating a task was updated.
func (t *Task) Updated(ctx context.Context, task internal.Task) error {
	return t.publish(ctx, "Task.Updated", "tasks.event.updated", task)
}

func (t *Task) publish(ctx context.Context, spanName, msgType string, task internal.Task) error {
	_, span := otel.Tracer(otelName).Start(ctx, spanName)
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   semconv.MessagingSystemKey,
			Value: attribute.StringValue("kafka"),
		},
		attribute.KeyValue{
			Key:   semconv.MessagingDestinationKey,
			Value: attribute.StringValue(t.topicName),
		},
	)

	//-

	var b bytes.Buffer

	evt := event{
		Type:  msgType,
		Value: task,
	}

	if err := json.NewEncoder(&b).Encode(evt); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "json.Encode")
	}

	if err := t.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &t.topicName,
			Partition: kafka.PartitionAny,
		},
		Value: b.Bytes(),
	}, nil); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "product.Producer")
	}

	return nil
}
