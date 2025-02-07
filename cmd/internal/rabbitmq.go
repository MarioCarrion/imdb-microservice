package internal

import (
	"github.com/streadway/amqp"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
)

type // RabbitMQ ...
RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewRabbitMQ instantiates the RabbitMQ instances using configuration defined in environment variables.
func NewRabbitMQ(conf *envvar.Configuration) (*RabbitMQ, error) {
	// XXX: Instead of using `RABBITMQ_URL` perhaps it makes sense to define
	// concrete `RABBIT_XYZ` variables similar to what we do for PostgreSQL to
	// clearly separate each field, like hostname, username, password, etc.
	url, err := conf.Get("RABBITMQ_URL")
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "conf.Get")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "amqp.Dial")
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "conn.Channel")
	}

	err = channel.ExchangeDeclare(
		"tasks", // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ch.ExchangeDeclare")
	}

	if err := channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ch.Qos")
	}

	// XXX: Dead Letter Exchange will be implemented in future episodes

	return &RabbitMQ{
		Connection: conn,
		Channel:    channel,
	}, nil
}

// Close ...
func (r *RabbitMQ) Close() error {
	if err := r.Connection.Close(); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "Connection.Close")
	}

	return nil
}
