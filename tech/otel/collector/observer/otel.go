// Package observer implements an OpenTelemetry Collector receiver that
package observer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"google.golang.org/grpc"
)

// newResource creates a new resource describing this application.
func newResource(ctx context.Context) (*resource.Resource, error) {
	// res: resource in opentelemetry. `resource` should embeded into service telemetry data: logs, metrics, traces
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
			semconv.DeploymentEnvironmentName("production"),
			attribute.String("language", "go"),
			attribute.String("author", "phamnam2003"), // custom attribute, this attribute should embeded into each query. It make other people easy to know who create this service
			attribute.StringSlice("contributors", []string{"chatgpt", "claud.ai", "deepseek"}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// merge with default resource, this make sure no missing attribute in resource
	res, err = resource.Merge(resource.Default(), res)
	if err != nil {
		if errors.Is(err, resource.ErrPartialResource) || errors.Is(err, resource.ErrSchemaURLConflict) {
			log.Printf("warning: partial resource merged: %v", err)
		}
		return nil, fmt.Errorf("failed to merged resource: %w", err)
	}

	return res, nil
}

// newPropagator creates a composite propagator that supports W3C Trace Context and Baggage.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

// SetupOtelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOtelSDK(ctx context.Context, conn *grpc.ClientConn) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	res, err := newResource(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	// Set up tp: trace provider.
	tp, err := newTraceProvider(ctx, conn, res)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	// Set up meter provider.
	mp, err := newMeterProvider(ctx, conn, res)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	// Setup logger provider.
	lp, err := newLoggerProvider(ctx, conn, res)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, lp.Shutdown)
	global.SetLoggerProvider(lp)

	return shutdown, err
}
