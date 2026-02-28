package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	brokers = flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka brokers")
	topic   = flag.String("topic", "otel-kafka-demo", "Kafka topic to produce messages to")
)

func newTracerProvider(rs *resource.Resource) (*tracesdk.TracerProvider, error) {
	// create new tracer provider
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable create trace exporter: %v", err)
	}

	// create a new batch span processor
	bsp := tracesdk.NewBatchSpanProcessor(exporter)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSpanProcessor(bsp),
		tracesdk.WithResource(rs),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)
	return tp, nil
}

func newMeterProvider(rs *resource.Resource) (*metric.MeterProvider, error) {
	// create a new meter provider
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, fmt.Errorf("unable create metric exporter: %v", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithResource(rs),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)
	return mp, nil
}

func newKotelTracer(tp *tracesdk.TracerProvider) *kotel.Tracer {
	// Create a new kotel tracer with the provided tracer provider and propagator.
	tracerOpts := []kotel.TracerOpt{
		kotel.TracerProvider(tp),
		kotel.TracerPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})),
	}
	return kotel.NewTracer(tracerOpts...)
}

func newKotelMeter(mp *metric.MeterProvider) *kotel.Meter {
	// Create a new kotel meter using a provided meter provider.
	meterOpts := []kotel.MeterOpt{
		kotel.MeterProvider(mp),
	}
	return kotel.NewMeter(meterOpts...)
}

func newKotel(tracer *kotel.Tracer, meter *kotel.Meter) *kotel.Kotel {
	kotelOpts := []kotel.Opt{
		kotel.WithTracer(tracer),
		kotel.WithMeter(meter),
	}
	return kotel.NewKotel(kotelOpts...)
}

func newProducerClient(kotelSvc *kotel.Kotel) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
		kgo.WithHooks(kotelSvc.Hooks()...),
		kgo.AllowAutoTopicCreation(),
	}
	return kgo.NewClient(opts...)
}

func produceMsg(client *kgo.Client, tracer trace.Tracer) error {
	// Start a new span with options.
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes([]attribute.KeyValue{attribute.String("some-key", "value")}...),
	}
	ctx, span := tracer.Start(context.Background(), "req", opts...)
	// End the span when function exits.
	defer span.End()

	// simulate some work by sleeping for 1 second.
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: *topic, Value: []byte("some-value"), Key: []byte("some-key")}
	// Pass in the context from the tracer.Start() call to ensure that the span
	// created is linked to the parent span.
	client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			log.Printf("error producing message: %v", err)
			// Set the status and record error on the span.
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	})

	wg.Wait()
	return nil
}

func newConsumerClient(kotelSvc *kotel.Kotel) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
		kgo.WithHooks(kotelSvc.Hooks()...),
		kgo.ConsumeTopics(*topic),
	}
	return kgo.NewClient(opts...)
}

func processRecord(record *kgo.Record, tracer *kotel.Tracer) {
	_, span := tracer.WithProcessSpan(record)
	// simulate some work by sleeping for 1 second.
	time.Sleep(1 * time.Second)
	// end the span after processing the record.
	defer span.End()

	log.Printf("consumed record with key: %s, value: %s, topic: %s, partition: %d, offset: %d",
		string(record.Key),
		string(record.Value),
		record.Topic,
		record.Partition,
		record.Offset,
	)
}

func consumeMsgs(client *kgo.Client, tracer *kotel.Tracer) error {
	fs := client.PollFetches(context.Background())
	if errs := fs.Errors(); len(errs) > 0 {
		return fmt.Errorf("error fetching messages: %v", errs)
	}

	iter := fs.RecordIter()
	for !iter.Done() {
		record := iter.Next()
		processRecord(record, tracer)
	}

	return nil
}

func do() error {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("otel-kafka-demo"),
			semconv.ServiceNamespaceKey.String("tech.kafka.ns_plugin_kotel"),
			semconv.ServiceVersionKey.String("0.1.0"),
			semconv.ServiceInstanceIDKey.String(uuid.NewString()),
			semconv.DeploymentEnvironmentName("dev"),
		),
	)
	if err != nil {
		return fmt.Errorf("could not set resources: %v", err)
	}
	tracerProvider, err := newTracerProvider(res)
	if err != nil {
		return err
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Printf("error shutdown trace provider: %v", err)
		}
	}()

	// Initialize meter provider and handle shutdown.	// Initialize meter provider and handle shutdown.
	meterProvider, err := newMeterProvider(res)
	if err != nil {
		return fmt.Errorf("could not initialize meter provider: %v", err)
	}
	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Printf("error shutdown meter provider: %v", err)
		}
	}()

	// Create a new kotel tracer and meter.
	kotelTracer := newKotelTracer(tracerProvider)
	kotelMeter := newKotelMeter(meterProvider)

	// Create a new kotel service.
	kotelSvc := newKotel(kotelTracer, kotelMeter)

	// Initialize producer client and handle close.
	producerClient, err := newProducerClient(kotelSvc)
	if err != nil {
		return fmt.Errorf("could not create producer client with kotel service: %v", err)
	}
	defer producerClient.Close()

	// Create request tracer and produce message.
	reqTracer := tracerProvider.Tracer("req-tracer")
	if err := produceMsg(producerClient, reqTracer); err != nil {
		return fmt.Errorf("could not produce message: %v", err)
	}

	// Initialize consumer client and handle close.
	consumerClient, err := newConsumerClient(kotelSvc)
	if err != nil {
		return fmt.Errorf("could not create consumer client with kotel service: %v", err)
	}
	defer consumerClient.Close()

	// Pass in the kotel tracer and consume messages in a loop.
	for {
		if err := consumeMsgs(consumerClient, kotelTracer); err != nil {
			return fmt.Errorf("error consuming messages: %v", err)
		}
	}
}

func main() {
	flag.Parse()

	if err := do(); err != nil {
		log.Fatalf("program crash with error: %v", err)
	}
}
