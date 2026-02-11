package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Order struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Amount    int    `json:"amount"`
	ProductID string `json:"product_id"`
}

func JitterBackoff(base time.Duration, factor float64) time.Duration {
	delay := time.Duration(float64(base) * factor)
	return base + time.Duration(rand.Int64N(int64(2*delay))) - delay
}

func main() {
	producer, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.ClientID("simply-kafka-producer"),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordRetries(10),
		kgo.RetryTimeout(30*time.Second),
		kgo.RetryBackoffFn(func(i int) time.Duration {
			return JitterBackoff(10*time.Millisecond, float64(i)*0.01)
		}),
		kgo.ProducerLinger(5*time.Millisecond),
		kgo.ProducerBatchMaxBytes(1<<20), // 1 MB
		kgo.MaxBufferedRecords(10_000),
		kgo.MaxBufferedBytes(256<<20), // 256 MB
		// compression available: None, Gzip, Snappy, Lz4, Zstd.
		// The Best of high workload is Zstd, best to balance CPU and network is Snappy
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, func() string {
			return "[kgo-kafka-producer]"
		})),
	)
	if err != nil {
		log.Fatal("failed create kafka producer: ", err)
	}
	defer producer.Close()

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.ClientID("simply-kafka-consumer"),
		kgo.ConsumerGroup("simply-kafka-group"),
		kgo.ConsumeTopics("simply-kafka-topic"),
		kgo.DisableAutoCommit(),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.FetchMaxPartitionBytes(1<<20),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxWait(500*time.Millisecond),
		kgo.SessionTimeout(45*time.Second),
		kgo.HeartbeatInterval(3*time.Second),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, func() string {
			return "[kgo-kafka-consumer]"
		})),
	)
	if err != nil {
		log.Fatal("failed create kafka consumer: ", err)
	}
	defer consumer.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	producer.Flush(ctx)

	go func() {
		orders := []Order{
			{ID: uuid.NewString(), UserID: uuid.NewString(), Amount: 100, ProductID: uuid.NewString()},
			{ID: uuid.NewString(), UserID: uuid.NewString(), Amount: 200, ProductID: uuid.NewString()},
			{ID: uuid.NewString(), UserID: uuid.NewString(), Amount: 150, ProductID: uuid.NewString()},
		}

		for _, o := range orders {
			value, _ := json.Marshal(o)
			rec := &kgo.Record{
				Topic: "simply-kafka-topic",
				Key:   []byte(o.ID),
				Value: value,
			}

			producer.Produce(context.Background(), rec, func(r *kgo.Record, err error) {
				if err != nil {
					log.Println("failed to produce record:", err)
					return
				}
				log.Println("produced record with ID:", o.ID)
			})
		}
	}()

	go func() {
		for {
			if ctx.Err() != nil {
				log.Println("context canceled, stopping consumer")
				return
			}
			fetches := consumer.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				log.Println("fetch errors:", errs)
				continue
			}

			fetches.EachRecord(func(r *kgo.Record) {
				var o Order
				if err := json.Unmarshal(r.Value, &o); err != nil {
					log.Println("failed to unmarshal record:", err)
					return
				}

				// commits after handle event from kafka
				if err := consumer.CommitRecords(ctx, r); err != nil {
					log.Println("failed to commit record:", err)
				}
				log.Printf("committed record with ID: %s, UserID: %s, Amount: %d, ProductID: %s\n", o.ID, o.UserID, o.Amount, o.ProductID)
			})
		}
	}()

	<-ctx.Done()
}
