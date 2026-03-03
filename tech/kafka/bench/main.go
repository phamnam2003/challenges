package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
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
	poolProduce         = flag.Bool("pool", false, "if true, use a sync.Pool to reuse record structs/slices (producing)")

	psync     = flag.Bool("psync", false, "produce synchronously")
	batchRecs = flag.Int("batch-recs", 1, "number of records to create before produce calls")
	pgoros    = flag.Int("pgoros", 1, "number of goroutines concurrently spawn to produce")

	caFile   = flag.String("ca-cert", "", "if non-empty, path to CA cert to use for TLS (implies -tl")
	certFile = flag.String("client-cert", "", "if non-empty, path to client cert to use for TLS (requires -client-key, implies -tls)")
	keyFile  = flag.String("client-key", "", "if non-empty, path to client key to use for TLS (requires -client-cert, implies -tls)")
	dialTLS  = flag.Bool("tls", false, "if true, use TLS to connect to brokers")

	consume = flag.Bool("consume", false, "if true, consume rather than produce")
	group   = flag.String("group", "bench-franz-go-group", "if non-empty, group to use for consuming rather than direct partition consuming (consuming)")

	logLevel = flag.String("log-level", "", "if non-empty, the log level to use (debug, info, warn, error)")

	saslMethod = flag.String("sasl-method", "", "if non-empty, sasl method to use (must specify all opts; spports plain, scram-sha-256, scram-sha-512, aws-msk-iam)")
	saslUser   = flag.String("sasl-user", "", "if non-empty, the sasl username to use")
	saslPass   = flag.String("sasl-pass", "", "if non-empty, the sasl password to use")

	rateRecs    int64
	rateBytes   int64
	staticValue []byte
	staticPool  = sync.Pool{
		New: func() any {
			return kgo.SliceRecord(staticValue)
		},
	}
	p = sync.Pool{
		New: func() any {
			return kgo.SliceRecord(make([]byte, *recordBytes))
		},
	}
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

func newRecord(num int64) *kgo.Record {
	var r *kgo.Record
	if *useStaticValue {
		return staticPool.Get().(*kgo.Record)
	} else if *poolProduce {
		r = p.Get().(*kgo.Record)
	} else {
		r = kgo.SliceRecord(make([]byte, *recordBytes))
	}
	formatValue(num, r.Value)

	return r
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
		kgo.AllowAutoTopicCreation(),
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

	if *saslMethod != "" || *saslUser != "" || *saslPass != "" {
		if *saslMethod == "" || *saslUser == "" || *saslPass == "" {
			log.Fatal("all of -sasl-method, -sasl-user, -sasl-pass must be specified if any are")
		}

		method := strings.ToLower(*saslMethod)
		method = strings.ReplaceAll(method, "-", "")
		method = strings.ReplaceAll(method, "_", "")
		switch method {
		case "plain":
			opts = append(opts, kgo.SASL(plain.Auth{
				User: *saslUser,
				Pass: *saslPass,
			}.AsMechanism()))
		case "scramsha256":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: *saslUser,
				Pass: *saslPass,
			}.AsSha256Mechanism()))
		case "scramsha512":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: *saslUser,
				Pass: *saslPass,
			}.AsSha512Mechanism()))
		case "awsmskiam":
			opts = append(opts, kgo.SASL(aws.Auth{
				AccessKey: *saslUser,
				SecretKey: *saslPass,
			}.AsManagedStreamingIAMMechanism()))
		default:
			log.Fatalf("unsupported sasl method: %s", *saslMethod)
		}
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable create kgo client: %v", err)
	}

	if *pprofPort != "" {
		go func() {
			err := http.ListenAndServe(*pprofPort, nil)
			if err != nil {
				log.Fatalf("pprof server failed: %v", err)
			}
		}()
	}

	go printRate()

	switch {
	case *consume:
		for {
			fs := cl.PollFetches(context.Background())
			fs.EachError(func(topic string, partition int32, err error) {
				if err != nil {
					log.Fatalf("error in fetch for topic %s partition %d: %v", topic, partition, err)
				}
			})
			var recs, bytes int64
			fs.EachRecord(func(r *kgo.Record) {
				recs++
				bytes += int64(len(r.Value))
			})
			atomic.AddInt64(&rateRecs, recs)
			atomic.AddInt64(&rateBytes, bytes)
		}

	case !*psync:
		var num atomic.Int64
		for range *pgoros {
			go func() {
				var recs []*kgo.Record
				for {
					recs = recs[:0]
					for range *batchRecs {
						recs = append(recs, newRecord(num.Add(1)))
					}
					for _, r := range recs {
						cl.Produce(context.Background(), r, func(r *kgo.Record, err error) {
							if *useStaticValue {
								staticPool.Put(r)
							} else if *poolProduce {
								p.Put(r)
							}
							if err != nil {
								log.Fatalf("produce record failed: %v", err)
							}
							atomic.AddInt64(&rateRecs, 1)
							atomic.AddInt64(&rateBytes, int64(*recordBytes))
						})
					}
				}
			}()
		}
		select {}

	default:
		var num atomic.Int64
		for range *pgoros {
			go func() {
				var recs []*kgo.Record
				for {
					recs = recs[:0]
					for range *batchRecs {
						recs = append(recs, newRecord(num.Add(1)))
					}

					ress := cl.ProduceSync(context.Background(), recs...)
					go func() {
						for _, res := range ress {
							r, err := res.Record, res.Err
							if *useStaticValue {
								staticPool.Put(r)
							} else if *poolProduce {
								p.Put(r)
							}

							if err != nil {
								log.Fatalf("produce error: %v", err)
							}
							atomic.AddInt64(&rateRecs, 1)
							atomic.AddInt64(&rateBytes, int64(*recordBytes))
						}
					}()
				}
			}()
		}
		select {}

	}
}
