package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/MarioCarrion/todo-api/cmd/internal"
	internaldomain "github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/elasticsearch"
	"github.com/MarioCarrion/todo-api/internal/envvar"
)

const rabbitMQConsumerName = "elasticsearch-indexer"

func main() {
	var env string

	flag.StringVar(&env, "env", "", "Environment Variables filename")
	flag.Parse()

	errC, err := run(env)
	if err != nil {
		log.Fatalf("Couldn't run: %s", err)
	}

	if err := <-errC; err != nil {
		log.Fatalf("Error while running: %s", err)
	}
}

func run(env string) (<-chan error, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "zap.NewProduction")
	}

	if err := envvar.Load(env); err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "envvar.Load")
	}

	vault, err := internal.NewVaultProvider()
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewVaultProvider")
	}

	conf := envvar.New(vault)

	//-

	esClient, err := internal.NewElasticSearch(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewElasticSearch")
	}

	rmq, err := internal.NewRabbitMQ(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.newRabbitMQ")
	}

	//-

	_, err = internal.NewOTExporter(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "newOTExporter")
	}

	//-

	srv := &Server{
		logger: logger,
		rmq:    rmq,
		task:   elasticsearch.NewTask(esClient),
		done:   make(chan struct{}),
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-ctx.Done()

		logger.Info("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer func() {
			_ = logger.Sync()

			rmq.Close()
			stop()
			cancel()
			close(errC)
		}()

		if err := srv.Shutdown(ctxTimeout); err != nil { //nolint: contextcheck
			errC <- err
		}

		logger.Info("Shutdown completed")
	}()

	go func() {
		logger.Info("Listening and serving")

		if err := srv.ListenAndServe(); err != nil {
			errC <- err
		}
	}()

	return errC, nil
}

type Server struct {
	logger *zap.Logger
	rmq    *internal.RabbitMQ
	task   *elasticsearch.Task
	done   chan struct{}
}

// ListenAndServe ...
// XXX: Dead Letter Exchange will be implemented in future episodes.
func (s *Server) ListenAndServe() error {
	queue, err := s.rmq.Channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "channel.QueueDeclare")
	}

	err = s.rmq.Channel.QueueBind(
		queue.Name,      // queue name
		"tasks.event.*", // routing key
		"tasks",         // exchange
		false,
		nil,
	)
	if err != nil {
		return internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "channel.QueueBind")
	}

	msgs, err := s.rmq.Channel.Consume(
		queue.Name,           // queue
		rabbitMQConsumerName, // consumer
		false,                // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
	if err != nil {
		return internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "channel.Consume")
	}

	go func() {
		for msg := range msgs {
			s.logger.Info("Received message: " + msg.RoutingKey)

			var nack bool

			// XXX: Instrumentation to be added in a future episode

			// XXX: We will revisit defining these topics in a better way in future episodes
			switch msg.RoutingKey {
			case "tasks.event.updated", "tasks.event.created":
				task, err := decodeTask(msg.Body)
				if err != nil {
					return
				}

				if err := s.task.Index(context.Background(), task); err != nil {
					nack = true
				}
			case "tasks.event.deleted":
				id, err := decodeID(msg.Body)
				if err != nil {
					return
				}

				if err := s.task.Delete(context.Background(), id); err != nil {
					nack = true
				}
			default:
				nack = true
			}

			if nack {
				s.logger.Info("NAcking :(")

				_ = msg.Nack(false, nack)
			} else {
				s.logger.Info("Acking :)")

				_ = msg.Ack(false)
			}
		}

		s.logger.Info("No more messages to consume. Exiting.")

		s.done <- struct{}{}
	}()

	return nil
}

// Shutdown ...
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	_ = s.rmq.Channel.Cancel(rabbitMQConsumerName, false)

	for {
		select {
		case <-ctx.Done():
			return internaldomain.WrapErrorf(ctx.Err(), internaldomain.ErrorCodeUnknown, "context.Done")
		case <-s.done:
			return nil
		}
	}
}

func decodeTask(b []byte) (internaldomain.Task, error) {
	var res internaldomain.Task

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&res); err != nil {
		return internaldomain.Task{}, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "gob.Decode")
	}

	return res, nil
}

func decodeID(b []byte) (string, error) {
	var res string

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&res); err != nil {
		return "", internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "gob.Decode")
	}

	return res, nil
}
