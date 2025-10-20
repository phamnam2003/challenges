package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newTraceProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	conn, err := grpc.NewClient("0.0.0.0:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// res: resource in opentelemetry
	res, err := resource.New(ctx,
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),         // Discover and provide host information.
		resource.WithAttributes(
			semconv.ServiceName("otel-http-demo"),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
			attribute.String("language", "go"),
		),
	)
	if err != nil {
		if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
			log.Printf("warning: partial resource created: %v", err)
		}
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
			log.Printf("warning: partial resource merged: %v", err)
		}
		return nil, fmt.Errorf("failed to merged resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))),
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

func complexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tr := otel.Tracer("complex-tracer")

	// 🧳 Add baggage
	member, _ := baggage.NewMember("user.id", "67890")
	bg, _ := baggage.New(member)
	ctx = baggage.ContextWithBaggage(ctx, bg)

	// root span request
	ctx, rootspan := tr.Start(ctx, "complex-request-handler")
	defer rootspan.End()

	rootspan.SetAttributes(
		semconv.HTTPRequestMethodKey.String(r.Method),
		semconv.HTTPRouteKey.String("/complex"),
		attribute.String("handler.type", "complex"),
	)

	// Các spans cùng cấp - xử lý các tác vụ độc lập
	_, validationSpan := tr.Start(ctx, "validate-input")
	// Giả lập validation
	time.Sleep(50 * time.Millisecond)
	validationSpan.SetAttributes(attribute.Bool("input.valid", true))
	validationSpan.End()

	_, authSpan := tr.Start(ctx, "authenticate-user")
	// Giả lập xác thực
	time.Sleep(100 * time.Millisecond)
	authSpan.SetAttributes(attribute.String("user.role", "admin"))
	authSpan.End()

	// Span lồng nhau - xử lý dữ liệu với nhiều bước
	dataCtx, dataProcessingSpan := tr.Start(ctx, "process-data")
	dataProcessingSpan.AddEvent("started data processing")

	// Spans con bên trong xử lý dữ liệu
	_, dbSpan := tr.Start(dataCtx, "database-query")
	time.Sleep(150 * time.Millisecond) // Giả lập truy vấn DB
	dbSpan.SetAttributes(
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)
	dbSpan.End()

	_, cacheSpan := tr.Start(dataCtx, "cache-operation")
	time.Sleep(30 * time.Millisecond) // Giả lập cache
	cacheSpan.SetAttributes(
		attribute.String("cache.operation", "SET"),
		attribute.String("cache.key", "user:67890"),
	)
	cacheSpan.End()

	dataProcessingSpan.AddEvent("completed data processing")
	dataProcessingSpan.End()

	// Thêm một span cùng cấp khác
	_, formattingSpan := tr.Start(ctx, "format-response")
	time.Sleep(40 * time.Millisecond)
	formattingSpan.SetAttributes(attribute.String("response.format", "json"))
	formattingSpan.End()

	// Lấy thông tin từ baggage
	userID := baggage.FromContext(ctx).Member("user.id").Value()

	response := fmt.Sprintf(`{
		"status": "success",
		"message": "Complex operation completed",
		"user_id": "%s",
		"timestamp": "%s"
	}`, userID, time.Now().Format(time.RFC3339))

	rootspan.AddEvent("sending complex response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func main() {
	ctx := context.Background()

	// Khởi tạo tracer provider
	tp, err := newTraceProvider(ctx)
	if err != nil {
		log.Fatal("cannot create tracer provider", err)
	}

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Error shutting down tracer provider: ", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(helloHandler), "HelloHandler"))
	mux.Handle("/complex", otelhttp.NewHandler(http.HandlerFunc(complexHandler), "ComplexHandler"))

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("🚀 Server started at :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error: ", err)
		}
	}()

	// Chờ tín hiệu shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}
