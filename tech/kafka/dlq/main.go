// This example demonstrates a dead letter queue (DLQ) pattern: records that
// fail processing after retries are forwarded to a separate DLQ topic with
// error metadata in headers. A DLQ consumer can later inspect and reprocess
// these failed records.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	group            = "main_group"
	topic            = "main_topic"
	dlqTopic         = "dlq_topic"
	maxRecRetryCount = 3
)

type message struct {
	Topic       string    `json:"topic"`
	Key         []byte    `json:"key"`
	Value       []byte    `json:"value"`
	TimeStamp   time.Time `json:"timestamp"`
	Offset      int64     `json:"offset"`
	Partition   int32     `json:"partition"`
	LeaderEpoch int32     `json:"leader_epoch"`
}

type kafka struct {
	producer, consumer *kgo.Client
}

func (k *kafka) connectProducer() error {
	var err error

	k.producer, err = kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.AllowAutoTopicCreation(),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, func() string {
			return "[kgo-producer]"
		})),
	)
	if err != nil {
		return err
	}
	if err = k.producer.Ping(context.Background()); err != nil {
		return err
	}

	return err
}

func (k *kafka) connectConsumer() error {
	var err error

	k.consumer, err = kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, func() string {
			return "[kgo-consumer]"
		})),
	)
	if err != nil {
		return err
	}
	if err = k.consumer.Ping(context.Background()); err != nil {
		return err
	}

	return err
}

func (k *kafka) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fetches := k.consumer.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}
			if fetches.Empty() {
				continue
			}

			fetches.EachError(func(topic string, partition int32, err error) {
				log.Fatalf("topic %s partition %d have error: %v", topic, partition, err)
			})

			fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
				// Note: please look at the goroutine-per-partition examples
				// if you want better concurrency.
				for _, rec := range ftp.Records {
					if err := k.process(rec); err != nil {
						// Before DLQ we probably want to add a retry mechanism
						// to make sure that there's no obvious errors.
						if err = k.retry(rec); err != nil {
							failed, _ := json.Marshal(&message{
								Topic:       rec.Topic,
								Key:         rec.Key,
								Value:       rec.Value,
								TimeStamp:   rec.Timestamp,
								Offset:      rec.Offset,
								Partition:   rec.Partition,
								LeaderEpoch: rec.LeaderEpoch,
							})

							k.producer.Produce(ctx, &kgo.Record{
								Topic:     dlqTopic,
								Key:       rec.Key,
								Value:     failed,
								Timestamp: time.Now(),
								Headers: []kgo.RecordHeader{
									{Key: "status", Value: []byte("review")},
									{Key: "error", Value: []byte(err.Error())},
								},
							}, nil)
						}
					}
				}
			})
		}
	}
}

func (k *kafka) process(rec *kgo.Record) error {
	time.Sleep(1 * time.Second)

	msg := &message{
		Topic:       rec.Topic,
		Key:         rec.Key,
		Value:       rec.Value,
		TimeStamp:   rec.Timestamp,
		Offset:      rec.Offset,
		Partition:   rec.Partition,
		LeaderEpoch: rec.LeaderEpoch,
	}

	if rand.IntN(100)%2 != 0 {
		// fake error
		return errors.New("failed to process record")
	}
	log.Printf("processing message: %+v", msg)

	return nil
}

func (k *kafka) retry(rec *kgo.Record) error {
	var err error
	for i := 1; i <= maxRecRetryCount; i++ {
		if err = k.process(rec); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return nil
		}
	}
	return err
}

func (k *kafka) close() {
	k.producer.Close()
	k.consumer.Close()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	k := &kafka{}

	var wg sync.WaitGroup
	if err := k.connectProducer(); err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	if err := k.connectConsumer(); err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}

	for i := range 10 {
		k.producer.Produce(ctx, &kgo.Record{
			Topic:     topic,
			Key:       []byte("key" + strconv.Itoa(i)),
			Value:     []byte("value" + strconv.Itoa(i)),
			Timestamp: time.Now(),
		}, nil)
	}
	wg.Add(1)
	go k.run(ctx, &wg)

	<-ctx.Done()
	k.close()
	wg.Wait()
}
