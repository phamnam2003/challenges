package observer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var (
	zapOTel *zap.Logger
	once    sync.Once
)

// newLoggerProvider creates a new OpenTelemetry LoggerProvider that exports logs
func newLoggerProvider(ctx context.Context, conn *grpc.ClientConn, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn), otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
		Enabled:         true,
		InitialInterval: 5 * time.Second,
		MaxInterval:     30 * time.Second,
		MaxElapsedTime:  60 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	batchProcessor := sdklog.NewBatchProcessor(exporter, sdklog.WithExportTimeout(30*time.Second))
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(batchProcessor),
		sdklog.WithResource(res),
	)
	setZapOTel(zapcore.InfoLevel, lp, os.Stdout)
	return lp, nil
}

// setZapOTel sets up a zap logger that writes to the provided writer and uses the given LoggerProvider
func setZapOTel(level zapcore.Level, lp *sdklog.LoggerProvider, writers ...io.Writer) {
	once.Do(func() {
		cores := make([]zapcore.Core, 0, len(writers)+1)
		for _, w := range writers {
			cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(w), level))
		}
		// add otelzap core bridges to push logs to OpenTelemetry Collector
		cores = append(cores, otelzap.NewCore("observer/otelzap/bridges", otelzap.WithSchemaURL(semconv.SchemaURL), otelzap.WithLoggerProvider(lp)))

		// NewTee combines multiple cores into one zap logger
		coreTee := zapcore.NewTee(
			cores...,
		)
		zapOTel = zap.New(coreTee, zap.AddCaller())
	})
}

// GetZapOTel returns the singleton zap logger configured for OpenTelemetry
func GetZapOTel() *zap.Logger {
	return zapOTel
}
