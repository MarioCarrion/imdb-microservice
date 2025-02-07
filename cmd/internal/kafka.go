package internal

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
)

// KafkaConsumer is the producer implementation of Kafka.
type KafkaProducer struct {
	Producer *kafka.Producer
	Topic    string
}

// NewKafkaProducer instantiates the Kafka producer using configuration defined in environment variables.
func NewKafkaProducer(conf *envvar.Configuration) (*KafkaProducer, error) {
	host, topic, err := newKafkaConfig(conf)
	if err != nil {
		return nil, internal.NewErrorf(internal.ErrorCodeUnknown, "newKafkaConfig")
	}

	config := kafka.ConfigMap{
		"bootstrap.servers": host,
	}

	client, err := kafka.NewProducer(&config)
	if err != nil {
		return nil, internal.NewErrorf(internal.ErrorCodeUnknown, "kafka.NewProducer")
	}

	return &KafkaProducer{
		Producer: client,
		Topic:    topic,
	}, nil
}

// KafkaConsumer is the consumer implementation of Kafka.
type KafkaConsumer struct {
	Consumer *kafka.Consumer
}

// NewKafkaConsumer instantiates the Kafka consumer using configuration defined in environment variables.
func NewKafkaConsumer(conf *envvar.Configuration, groupID string) (*KafkaConsumer, error) {
	host, topic, err := newKafkaConfig(conf)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "kafka.newKafkaConfig")
	}

	config := kafka.ConfigMap{
		"bootstrap.servers":  host,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	client, err := kafka.NewConsumer(&config)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "kafka.NewConsumer")
	}

	if err := client.Subscribe(topic, nil); err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "client.Subscribe")
	}

	return &KafkaConsumer{
		Consumer: client,
	}, nil
}

func newKafkaConfig(conf *envvar.Configuration) (host, topic string, err error) {
	host, err = conf.Get("KAFKA_HOST")
	if err != nil {
		return "", "", internal.WrapErrorf(err, internal.ErrorCodeUnknown, "conf.Get KAFKA_HOST")
	}

	topic, err = conf.Get("KAFKA_TOPIC")
	if err != nil {
		return "", "", internal.WrapErrorf(err, internal.ErrorCodeUnknown, "conf.Get KAFKA_TOPIC")
	}

	if topic == "" {
		return "", "", internal.NewErrorf(internal.ErrorCodeInvalidArgument, "KAFKA_TOPIC is required")
	}

	return host, topic, nil
}
