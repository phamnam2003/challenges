// Package observer implements an OpenTelemetry Collector receiver that
package observer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// NewTraceProvider creates a new OpenTelemetry TracerProvider that exports traces
func NewTraceProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// exporter to push into otel collector
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// merge with default resource, this make sure no missing attribute in resource
	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
			log.Printf("warning: partial resource merged: %v", err)
		}
		return nil, fmt.Errorf("failed to merged resource: %w", err)
	}

	// create new trace provider with exporter and resource, you need configure sampler with tail sampling.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(1.0)), // sampling get all traces, filter and tail-sampling configured in collector
		),
	)
	return tp, nil
}
