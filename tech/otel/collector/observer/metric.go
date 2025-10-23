package observer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
)

// newMeterProvider creates a new OpenTelemetry MeterProvider that exports metrics
func newMeterProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// metric exporter to push into otel collector
	meterExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn), otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
		Enabled:         true,
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  60 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// create meter provider with periodic reader (exporter with interval scrape duration) and resource
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(
			meterExporter,
			sdkmetric.WithInterval(10*time.Second), // export metrics every 10 seconds
		)),
		sdkmetric.WithResource(res),
	)
	return mp, nil
}
