package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"github.com/twmb/tlscfg"
)

var (
	brokers   = flag.String("brokers", "localhost:9092", "Comma separate list brokers to connect")
	topic     = flag.String("topic", "bench-franz-go-topic", "Topic name to produce messages to")
	pprofPort = flag.String("pprof", ":9876", "port to bind to for pprof, if non-empty")
	prom      = flag.Bool("prom", false, "if true, install a /metrics path for prometheus metrics to the default handler (usage requires -pprof)")

	useStaticValue = flag.Bool("static-record", false, "if true, use the same record value for every record (eliminates creating and formatting values for records; implies -pool)")

	recordBytes         = flag.Int("record-bytes", 100, "bytes per record value (producing)")
	batchMaxBytes       = flag.Int("batch-max-bytes", 1_000_000, "the maximum batch size to allow per-partition (must be less than Kafka's max.message.bytes, producing")
	compression         = flag.String("compression", "none", "the compression to use when producing: none, gzip, snappy, lz4, zstd")
	acks                = flag.Int("acks", -1, "acks required: 0, -1, 1")
	linger              = flag.Duration("linger", 0, "if non-zero, linger to use when producing")
	noIdempotency       = flag.Bool("no-idempotency", false, "if true, disable idempotency (force 1 produce rps)")
	noIdempotentMaxReqs = flag.Int("max-inflight-produce-per-broker", 5, "if idempotency is disabled, the number of produce requests to allow per broker")

	caFile   = flag.String("ca-cert", "", "if non-empty, path to CA cert to use for TLS (implies -tl")
	certFile = flag.String("client-cert", "", "if non-empty, path to client cert to use for TLS (requires -client-key, implies -tls)")
	keyFile  = flag.String("client-key", "", "if non-empty, path to client key to use for TLS (requires -client-cert, implies -tls)")
	dialTLS  = flag.Bool("tls", false, "if true, use TLS to connect to brokers")

	consume = flag.Bool("consume", false, "if true, consume rather than produce")
	group   = flag.String("group", "bench-franz-go-group", "if non-empty, group to use for consuming rather than direct partition consuming (consuming)")

	logLevel = flag.String("log-level", "", "if non-empty, the log level to use (debug, info, warn, error)")

	rateRecs    int64
	rateBytes   int64
	staticValue []byte
)

func printRate() {
	for range time.Tick(time.Second) {
		recs := atomic.SwapInt64(&rateRecs, 0)
		bytes := atomic.SwapInt64(&rateBytes, 0)
		log.Printf("%0.2f MiB/s, %0.2fk records/s", float64(bytes)/1024/1024, float64(recs)/1000)
	}
}

func formatValue(num int64, v []byte) {
	var buf [20]byte // max int64 takes 19 bytes, then we add a
	b := strconv.AppendInt(buf[:0], num, 10)
	b = append(b, ' ')

	n := copy(v, b)
	for n != len(v) {
		n += copy(v[n:], b)
	}
}

func main() {
	flag.Parse()

	var customTLS bool
	if *caFile != "" || *certFile != "" || *keyFile != "" {
		*dialTLS = true
		customTLS = true
	}

	if *recordBytes <= 0 {
		log.Fatal("record bytes must be larger than zero")
	}

	if *useStaticValue {
		staticValue = make([]byte, *recordBytes)
		formatValue(0, staticValue)
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.DefaultProduceTopic(*topic),
		kgo.MaxBufferedBytes(250<<20 / *recordBytes + 1),
		kgo.MaxConcurrentFetches(3),
		// We have good compression, so we want to limit what we read
		// back because snappy deflation will balloon our memory usage.
		kgo.FetchMaxBytes(5 << 20),
		kgo.ProducerBatchMaxBytes(int32(*batchMaxBytes)),
	}

	if *noIdempotency {
		opts = append(opts, kgo.DisableIdempotentWrite(), kgo.MaxProduceRequestsInflightPerBroker(*noIdempotentMaxReqs))
	}
	if *consume {
		opts = append(opts, kgo.ConsumeTopics(*topic), kgo.ConsumerGroup(*group))
	}

	switch *acks {
	case 0:
		opts = append(opts, kgo.RequiredAcks(kgo.NoAck()))
	case 1:
		opts = append(opts, kgo.RequiredAcks(kgo.LeaderAck()))
	default:
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	}

	if *prom {
		metrics := kprom.NewMetrics("kgo")
		http.Handle("/metrics", metrics.Handler())
		opts = append(opts, kgo.WithHooks(metrics))
	}

	switch strings.ToLower(*logLevel) {
	case "":
	case "debug":
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelDebug, nil)))
	case "info":
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)))
	case "warn":
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelWarn, nil)))
	case "error":
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelError, nil)))
	default:
		log.Fatalf("unknown log level: %s", *logLevel)
	}

	if *linger != 0 {
		opts = append(opts, kgo.ProducerLinger(*linger))
	}

	switch *compression {
	case "", "none":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.NoCompression()))
	case "gzip":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.GzipCompression()))
	case "snappy":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.SnappyCompression()))
	case "lz4":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.Lz4Compression()))
	case "zstd":
		opts = append(opts, kgo.ProducerBatchCompression(kgo.ZstdCompression()))
	default:
		log.Fatalf("unknown compression: %s", *compression)
	}

	if *dialTLS {
		if customTLS {
			tc, err := tlscfg.New(
				tlscfg.MaybeWithDiskCA(*caFile, tlscfg.ForClient),
				tlscfg.MaybeWithDiskKeyPair(*certFile, *keyFile),
			)
			if err != nil {
				log.Fatalf("unable to create TLS config: %v", err)
			}
			opts = append(opts, kgo.DialTLSConfig(tc))
		} else {
			opts = append(opts, kgo.DialTLSConfig(new(tls.Config)))
		}
	}
}
