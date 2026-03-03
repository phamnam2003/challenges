package main

import (
	"flag"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	brokers = flag.String("brokers", "localhost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	topic   = flag.String("topic", "bench-confluent-kafka-go-topic", "Kafka topic to produce to")

	recordBytes   = flag.Int("record-bytes", 100, "bytes per record (producing)")
	noCompression = flag.Bool("no-compression", false, "disable snappy compression (producing)")

	consume = flag.Bool("consume", false, "consume messages instead of producing")
	group   = flag.String("group", "bench-confluent-kafka-go-group", "if non-empty, group to use for consuming rather than direct partition consuming (consuming)")

	rateBytes, rateRecs int64
)

func printRate() {
	for range time.Tick(time.Second) {
		recs := atomic.SwapInt64(&rateRecs, 0)
		bytes := atomic.SwapInt64(&rateBytes, 0)
		log.Printf("Rate: %.2f MB/s, %.2f K recs/s", float64(bytes)/1024/1024, float64(recs)/1000)
	}
}

func newValue(num int64) []byte {
	var buf [20]byte // max int64 takes 19 bytes, then we add a space
	b := strconv.AppendInt(buf[:0], num, 10)
	b = append(b, ' ')

	s := make([]byte, *recordBytes)

	var n int
	for n != len(s) {
		n += copy(s[n:], b)
	}
	return s
}

func main() {
	flag.Parse()

	go printRate()

	switch *consume {
	case false:
		cfg := &kafka.ConfigMap{
			"bootstrap.servers":            *brokers,
			"enable.idempotence":           true,
			"queue.buffering.max.messages": 10000000,
			"linger.ms":                    10,
		}
		if *noCompression {
			(*cfg)["compression.codec"] = "none"
		}

		p, err := kafka.NewProducer(cfg)
		if err != nil {
			log.Fatalf("failed to create producer: %s", err)
		}

		go func() {
			for {
				switch ev := (<-p.Events()).(type) {
				case *kafka.Message:
					err := ev.TopicPartition.Error
					if err != nil {
						log.Fatalf("produce to topic partition error: %v", err)
					}
					atomic.AddInt64(&rateRecs, 1)
					atomic.AddInt64(&rateBytes, int64(*recordBytes))

				case kafka.Error:
					if ev.IsFatal() {
						log.Fatalf("fatal error: %v", ev)
					}
				}
			}
		}()

		var num int64
		for {
			err = p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
				Value:          newValue(num),
			}, nil)
			num++
			if err != nil {
				log.Fatalf("produce record confluent-kafka-go error: %v", err)
			}
		}

	case true:
		cfg := &kafka.ConfigMap{
			"bootstrap.servers": *brokers,
			"group.id":          *group,
			"auto.offset.reset": "earliest",
		}
		c, err := kafka.NewConsumer(cfg)
		if err != nil {
			log.Fatalf("failed to create consumer: %s", err)
		}

		err = c.Subscribe(*topic, nil)
		if err != nil {
			log.Fatalf("failed to subscribe to topic: %s", err)
		}

		for {
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				atomic.AddInt64(&rateRecs, 1)
				atomic.AddInt64(&rateBytes, int64(len(e.Value)))

			case kafka.Error:
				log.Fatalf("consume error: %v", e)

			}
		}
	}
}
