package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"

	"github.com/MarioCarrion/todo-api-microservice-example/cmd/internal"
	internaldomain "github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/elasticsearch"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar"
)

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

	kafka, err := internal.NewKafkaConsumer(conf, "elasticsearch-indexer")
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.NewKafkaConsumer")
	}

	//-

	_, err = internal.NewOTExporter(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "internal.newOTExporter ")
	}

	//-

	srv := &Server{
		logger: logger,
		kafka:  kafka,
		task:   elasticsearch.NewTask(esClient),
		doneC:  make(chan struct{}),
		closeC: make(chan struct{}),
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
			_ = kafka.Consumer.Unsubscribe()

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
	kafka  *internal.KafkaConsumer
	task   *elasticsearch.Task
	doneC  chan struct{}
	closeC chan struct{}
}

// ListenAndServe ...
func (s *Server) ListenAndServe() error {
	commit := func(msg *kafka.Message) {
		if _, err := s.kafka.Consumer.CommitMessage(msg); err != nil {
			s.logger.Error("commit failed", zap.Error(err))
		}
	}

	go func() {
		run := true

		for run {
			select {
			case <-s.closeC:
				run = false

				break
			default:
				msg, ok := s.kafka.Consumer.Poll(150).(*kafka.Message)
				if !ok {
					continue
				}

				var evt struct {
					Type  string
					Value internaldomain.Task
				}

				if err := json.NewDecoder(bytes.NewReader(msg.Value)).Decode(&evt); err != nil {
					s.logger.Info("Ignoring message, invalid", zap.Error(err))
					commit(msg)

					continue
				}

				ok = false

				switch evt.Type {
				case "tasks.event.updated", "tasks.event.created":
					if err := s.task.Index(context.Background(), evt.Value); err == nil {
						ok = true
					}
				case "tasks.event.deleted":
					if err := s.task.Delete(context.Background(), evt.Value.ID); err == nil {
						ok = true
					}
				}

				if ok {
					s.logger.Info("Consumed", zap.String("type", evt.Type))
					commit(msg)
				}
			}
		}

		s.logger.Info("No more messages to consume. Exiting.")

		s.doneC <- struct{}{}
	}()

	return nil
}

// Shutdown ...
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	close(s.closeC)

	for {
		select {
		case <-ctx.Done():
			return internaldomain.WrapErrorf(ctx.Err(), internaldomain.ErrorCodeUnknown, "context.Done")
		case <-s.doneC:
			return nil
		}
	}
}
