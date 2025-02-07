package memcached

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
)

const otelName = "github.com/MarioCarrion/todo-api-microservice-example/internal/memcached"

// XXX: "delete" and "set" intentionally ignore errors, a better approach
// would be to implement an unexported "client" type defining all the three
// methods defined in this file that includes a Circuit Breaker for logging
// errors and retry as needed.
// See https://youtu.be/UnL2iGcD7vE for more details about that pattern.

func deleteTask(ctx context.Context, client *memcache.Client, key string) {
	defer newOTELSpan(ctx, "deleteTask").End()

	//-

	_ = client.Delete(key)
}

func getTask(ctx context.Context, client *memcache.Client, key string, target interface{}) error {
	defer newOTELSpan(ctx, "getTask").End()

	//-

	item, err := client.Get(key)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "client.Get")
	}

	if err := gob.NewDecoder(bytes.NewReader(item.Value)).Decode(target); err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "gob.NewDecoder")
	}

	return nil
}

func setTask(ctx context.Context, client *memcache.Client, key string, value interface{}, expiration time.Duration) {
	defer newOTELSpan(ctx, "setTask").End()

	//-

	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(value); err != nil {
		return
	}

	_ = client.Set(&memcache.Item{
		Key:        key,
		Value:      b.Bytes(),
		Expiration: int32(time.Now().Add(expiration).Unix()), //nolint:gosec
	})
}

//-

func newOTELSpan(ctx context.Context, name string) trace.Span {
	_, span := otel.Tracer(otelName).Start(ctx, name)

	span.SetAttributes(semconv.DBSystemMemcached)

	return span
}
