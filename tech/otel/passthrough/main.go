package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const name = "otel.example/passthrough"

// nonGlobalTracer creates a trace provider instance for testing, but doesn't
// set it as the global tracer provider.
func nonGlobalTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stdouttrace exporter: %s", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)

	return tp, nil
}

func main() {
	ctx := context.Background()

	// init passthrough tracer with global scope
	// We explicitly DO NOT set the global TracerProvider using otel.SetTracerProvider().
	// The unset TracerProvider returns a "non-recording" span, but still passes through context.
	log.Println("Register a global TextMapPropagator, but do not register a global TracerProvider to be in \"passthrough\" mode.")
	log.Println("The \"passthrough\" mode propagates the TraceContext and Baggage, but does not record spans.")
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tp, err := nonGlobalTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = tp.Shutdown(ctx) }()

	// make an initial http request
	r, err := http.NewRequest("", "", http.NoBody)
	if err != nil {
		panic(err)
	}
	// This is roughly what an instrumented http client does.
	log.Println("The \"make outer request\" span should be recorded, because it is recorded with a Tracer from the SDK TracerProvider")

	var span trace.Span
	tracer := tp.Tracer(name)
	ctx, span = tracer.Start(ctx, "make outer request")
	defer span.End()

	r = r.WithContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	backendFunc := func(r *http.Request) {
		// This is roughly what an instrumented http server does.
		ctx := r.Context()
		tp := trace.SpanFromContext(ctx).TracerProvider()
		tracer := tp.Tracer(name)

		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))
		log.Println("The \"handle inner request\" span should be recorded, because it is recorded with a Tracer from the SDK TracerProvider")
		_, span := tracer.Start(ctx, "handle inner request")
		defer span.End()

		// Do "backend work"
		time.Sleep(time.Second)
	}

	// This handler will be a passthrough, since we didn't set a global TracerProvider
	passthroughHandler := NewHandler(backendFunc)
	passthroughHandler.HandleHTTPReq(r)
}
