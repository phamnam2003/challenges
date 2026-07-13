package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/phamnam2003/challenges/tech/otel/collector/observer"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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
		if err := shutdown(context.Background()); err != nil {
			log.Fatal("failed to shutdown otel sdk: ", err)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/complex", otelhttp.NewHandler(http.HandlerFunc(complexHandler), "ComplexHandler"))

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	m := otel.Meter("collector_metric")
	// create counter, histogram, gauge, updown metric
	// it is the same configuration as prometheus example in tech/otel/prometheus/main.go
	counter, err := m.Int64Counter("collector_requests_total")
	if err != nil {
		log.Fatal("failed to create counter: ", err)
	}
	counter.Add(ctx, rand.Int64N(120))

	gauge, err := m.Float64ObservableGauge("bar", metric.WithDescription("simple_gauge"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = m.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		n := -10. + rand.Float64()*(90.0) // [-10, 80)
		o.ObserveFloat64(gauge, n)
		return nil
	}, gauge)
	if err != nil {
		log.Fatal(err)
	}

	histogram, err := m.Float64Histogram(
		"baz",
		metric.WithDescription("a histogram with custom buckets and rename"),
		metric.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 136)
	histogram.Record(ctx, 64)
	histogram.Record(ctx, 701)
	histogram.Record(ctx, 830)

	observer.GetZapOTel().Info("Starting server at :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error: ", err)
	}
}
