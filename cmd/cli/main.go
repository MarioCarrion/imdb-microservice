package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"github.com/MarioCarrion/todo-api-microservice-example/pkg/openapi3"
)

func main() {
	initTracer()

	//-

	clientOA3 := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	client, err := openapi3.NewClientWithResponses("http://0.0.0.0:9234", openapi3.WithHTTPClient(&clientOA3))
	if err != nil {
		log.Fatalf("Couldn't instantiate client: %s", err)
	}

	newPtrStr := func(s string) *string {
		return &s
	}

	newPtrTime := func(t time.Time) *time.Time {
		return &t
	}

	count := 1

	for count < 101 {
		priority := openapi3.Low

		_, err := client.CreateTaskWithResponse(context.Background(),
			openapi3.CreateTaskJSONRequestBody{
				Dates: &openapi3.Dates{
					Start: newPtrTime(time.Now()),
					Due:   newPtrTime(time.Now().Add(time.Hour * 24)),
				},
				Description: newPtrStr(fmt.Sprintf("Searchable Task %d", count)),
				Priority:    &priority,
			})
		if err != nil {
			log.Fatalf("Couldn't create task %s", err)
		}

		count++
	}
}

func initTracer() {
	jaegerEndpoint := "http://localhost:14268/api/traces"

	jaegerExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
	)
	if err != nil {
		log.Fatalln("Couldn't initialize exporter", err)
	}

	// Create stdout exporter to be able to retrieve the collected spans.
	_, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalln("Couldn't initialize exporter", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(jaegerExporter),
		trace.WithResource(resource.NewSchemaless(attribute.KeyValue{
			Key:   semconv.ServiceNameKey,
			Value: attribute.StringValue("rest-server"),
		})),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}
