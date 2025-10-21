package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phamnam2003/challenges/tech/otel/collector/observer"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newMeterProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// create metrics exporter to push to otel collector with grpc connection
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// create meter provider with exporter and resource
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			exporter,
			sdkmetric.WithInterval(15*time.Second), // export metrics every 15 seconds
		)),
		sdkmetric.WithResource(res),
	)
	return mp, nil
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
	// this recommendation approach step with grpc connection, you can custom credentials with token (JWT, PASETO, etc.)
	conn, err := grpc.NewClient("0.0.0.0:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to create gRPC connection to collector: ", err)
	}

	shutdown, err := observer.SetupOtelSDK(ctx, conn)
	if err != nil {
		log.Fatal("failed to setup otel sdk: ", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown otel sdk: ", err)
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
}
