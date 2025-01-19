package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/MarioCarrion/todo-api/cmd/internal"
	internaldomain "github.com/MarioCarrion/todo-api/internal"
	"github.com/MarioCarrion/todo-api/internal/elasticsearch"
	"github.com/MarioCarrion/todo-api/internal/envvar"
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

	rdb, err := internal.NewRedis(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "newRedis")
	}

	//-

	_, err = internal.NewOTExporter(conf)
	if err != nil {
		return nil, internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "newOTExporter")
	}

	//-

	srv := &Server{
		logger: logger,
		rdb:    rdb,
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

			rdb.Close()
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
	rdb    *redis.Client
	pubsub *redis.PubSub
	task   *elasticsearch.Task
	done   chan struct{}
}

// ListenAndServe ...
func (s *Server) ListenAndServe() error {
	pubsub := s.rdb.PSubscribe(context.Background(), "tasks.*")

	_, err := pubsub.Receive(context.Background())
	if err != nil {
		return internaldomain.WrapErrorf(err, internaldomain.ErrorCodeUnknown, "pubsub.Receive")
	}

	s.pubsub = pubsub

	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			s.logger.Info("Received message: " + msg.Channel)

			// XXX: Instrumentation to be added in a future episode

			// XXX: We will revisit defining these topics in a better way in future episodes
			switch msg.Channel {
			case "tasks.event.updated", "tasks.event.created":
				var task internaldomain.Task

				if err := json.NewDecoder(strings.NewReader(msg.Payload)).Decode(&task); err != nil {
					s.logger.Info("Ignoring message, invalid", zap.Error(err))

					continue
				}

				if err := s.task.Index(context.Background(), task); err != nil {
					s.logger.Info("Couldn't index task", zap.Error(err))
				}

				s.logger.Info("Record saved")
			case "tasks.event.deleted":
				var id string

				if err := json.NewDecoder(strings.NewReader(msg.Payload)).Decode(&id); err != nil {
					s.logger.Info("Ignoring message, invalid", zap.Error(err))

					continue
				}

				if err := s.task.Delete(context.Background(), id); err != nil {
					s.logger.Info("Couldn't delete task", zap.Error(err))
				}

				s.logger.Info("Record deleted")
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

	s.pubsub.Close()

	for {
		select {
		case <-ctx.Done():
			return internaldomain.WrapErrorf(ctx.Err(), internaldomain.ErrorCodeUnknown, "context.Done")
		case <-s.done:
			return nil
		}
	}
}
