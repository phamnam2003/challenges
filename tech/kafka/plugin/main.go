// This example demonstrates franz-go's logging and metrics plugin ecosystem.
//
// Logging plugins (choose one):
//   - kslog: integrates with log/slog (recommended for new projects)
//   - kzap: integrates with go.uber.org/zap
//   - kgo.BasicLogger: built-in, no dependencies
//
// Metrics plugins (can be used alongside any logger):
//   - kprom: exports Prometheus metrics via an HTTP endpoint
//
// You can use one logger and one metrics plugin together. This example
// registers both a logger and Prometheus metrics to show how they combine.
// Use -logger to select the logging backend.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"github.com/twmb/franz-go/plugin/kslog"
	"github.com/twmb/franz-go/plugin/kzap"
	"go.uber.org/zap"
)

var (
	brokers     = flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka brokers")
	topic       = flag.String("topic", "otel-kafka-demo", "Kafka topic to produce messages to")
	producer    = flag.Bool("producer", false, "Run as producer")
	logType     = flag.String("logger", "slog", "Logging plugin to use (slog, kzap, basic)")
	metricsPort = flag.Int("metrics-port", 2112, "Port to expose Prometheus metrics on (if using kprom)")
)

func main() {
	flag.Parse()
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.DefaultProduceTopic(*topic),
	}

	// Configure a logging plugin. You only need one logger.
	switch *logType {
	case "slog":
		sl := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
		opts = append(opts, kgo.WithLogger(kslog.New(sl)))

	case "zap":
		z, err := zap.NewDevelopment()
		if err != nil {
			log.Fatalf("unable to create zap logger: %v", err)
		}
		opts = append(opts, kgo.WithLogger(kzap.New(z)))

	case "basic":
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)))

	default:
		log.Fatalf("unknown logger type: %s", *logType)
	}

	// Configure Prometheus metrics. Metrics hooks can be used alongside
	// any logger - they are independent. Use port 0 to skip metrics.
	if *metricsPort > 0 {
		metrics := kprom.NewMetrics("kgo")
		opts = append(opts, kgo.WithHooks(metrics))

		go func() {
			http.Handle("/metrics", metrics.Handler())
			log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", *metricsPort), nil))
		}()
		log.Printf("Prometheus metrics available at http://localhost:%d/metrics", *metricsPort)
	}

	if !*producer {
		opts = append(opts, kgo.ConsumeTopics(*topic))
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *producer {
		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(1 * time.Second):
			}

			if err := cl.ProduceSync(ctx, kgo.StringRecord("record testing with kafka plugin franz-go")); err != nil {
				log.Fatalf("cannot produce: %v", err)
			}
		}
	} else {
		for {
			fs := cl.PollFetches(ctx)
			if fs.IsClientClosed() {
				return
			}

			fs.EachError(func(t string, p int32, err error) {
				log.Fatalf("error in topic %s partition %d: %v", t, p, err)
			})

			var seen int
			fs.EachRecord(func(r *kgo.Record) {
				seen++
			})

			log.Printf("fetched %d records", seen)
		}
	}
}
