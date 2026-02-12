package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedBrokers = flag.String("brokers", "localhost:9092", "comma delimited list of seed brokers")
	topicName   = flag.String("topic", "manual-flushing-topic", "topic to produce to")
	batchSize   = flag.Int("batch-size", 100, "number of records to batch before flushing")
)

func main() {
	flag.Parse()
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...),
		kgo.DefaultProduceTopic(*topicName),
		kgo.AllowAutoTopicCreation(),
		// ManualFlushing disables automatic flushing. Records are
		// buffered until Flush is called. If MaxBufferedRecords is
		// reached before a Flush, Produce returns ErrMaxBuffered.
		kgo.ManualFlushing(),
		kgo.MaxBufferedRecords(*batchSize*2),
	)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	ctx := context.Background()

	for batch := 0; ; batch++ {
		// Buffer a batch of records without sending them.
		for i := range *batchSize {
			r := kgo.StringRecord(fmt.Sprintf("batch=%d record=%d", batch, i))
			cl.TryProduce(ctx, r, func(r *kgo.Record, err error) {
				if err != nil {
					log.Printf("record delivery error: %v", err)
				}
			})
		}
		log.Printf("buffered %d records (batch %d), flushing ...", cl.BufferedProduceRecords(), batch)
		// Flush sends all buffered records and waits for completion.
		if err := cl.Flush(ctx); err != nil {
			log.Fatalf("flush record to kafka error: %v", err)
		}
		log.Printf("batch %d flushed successfully", batch)

		time.Sleep(5 * time.Second)
	}
}
