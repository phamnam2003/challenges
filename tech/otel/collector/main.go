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
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newTraceProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	conn, err := grpc.NewClient("http://localhost:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	// res: resource of opentelemetry
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String("otel-http-demo"),
		attribute.String("env", "dev"),
		attribute.String("exporter", "otlp"),
	))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	return tp, nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 🧳 Add baggage
	member, _ := baggage.NewMember("user.id", "12345")
	bg, _ := baggage.New(member)
	ctx = baggage.ContextWithBaggage(ctx, bg)

	tr := otel.Tracer("example-tracer")
	_, span := tr.Start(ctx, "handle-hello")
	defer span.End()

	span.SetAttributes(semconv.HTTPRequestMethodGet)
	span.AddEvent("processing request")

	time.Sleep(200 * time.Millisecond) // simulate work

	userID := baggage.FromContext(ctx).Member("user.id").Value()
	msg := fmt.Sprintf("Hello! Baggage user.id=%s", userID)
	span.AddEvent("sending response")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg))
}

func main() {
	ctx := context.Background()

	// tp: tracer provider
	tp, err := newTraceProvider(ctx)
	if err != nil {
		log.Fatal("cannot create tracer provider", err)
	}
	defer tp.Shutdown(ctx)

	mux := http.NewServeMux()
	mux.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(helloHandler), "HelloHanlder"))

	log.Println("🚀 Server started at :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
