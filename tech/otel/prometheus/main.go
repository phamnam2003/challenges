package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "otel.example/prometheus"

func main() {
	ctx := context.Background()

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	defer provider.Shutdown(ctx)

	meter := provider.Meter(meterName)

	go serveMetrics()

	opt := metric.WithAttributes(
		attribute.String("environment", "production"),
		attribute.String("region", "us-west"),
		attribute.String("service", "payment-service"),
		attribute.String("version", "1.0.0"),
		attribute.String("instance_id", "instance-12345"),
		attribute.String("team", "backend"),
	)

	// Counter (chắc chắn hoạt động)
	counter, err := meter.Int64Counter("foo", metric.WithDescription("simple_counter"))
	if err != nil {
		log.Fatal(err)
	}
	counter.Add(ctx, rand.Int64N(120), opt)

	// CÁCH 1: Observable Gauge với giá trị động
	gauge, err := meter.Float64ObservableGauge("bar", metric.WithDescription("simple_gauge"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		n := -10. + rand.Float64()*(90.0) // [-10, 80)
		o.ObserveFloat64(gauge, n, opt)
		return nil
	}, gauge)
	if err != nil {
		log.Fatal(err)
	}

	histogram, err := meter.Float64Histogram(
		"baz",
		metric.WithDescription("a histogram with custom buckets and rename"),
		metric.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	<-ctx.Done()
}

func serveMetrics() {
	log.Printf("Serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2223", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
