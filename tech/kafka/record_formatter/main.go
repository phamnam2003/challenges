// This example demonstrates the RecordFormatter and RecordReader, which
// provide printf-style formatting for records. These are useful for building
// CLI tools, log processing, and format conversion.
//
// The formatter uses percent verbs:
//
// %t  topic          %T  topic length
// %k  key            %K  key length
// %v  value          %V  value length
// %p  partition      %o  offset
// %d  timestamp      %H  number of headers
// %h  header spec    %e  leader epoch
//
// See NewRecordFormatter documentation for the full specification.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers = flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka brokers")
	topic   = flag.String("topic", "record-formatter-topic", "Kafka topic to produce messages to")
	layout  = flag.String("format", "%t [p%p o%o] %d: key=%k value=%v headers=%H%h{ %k=%v}\n",
		"record format layout (see RecordFormatter docs)")
)

func main() {
	flag.Parse()

	formatter, err := kgo.NewRecordFormatter(*layout)
	if err != nil {
		log.Fatalf("unable to create record formatter: %v", err)
	}

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, nil)),
		kgo.DefaultProduceTopic(*topic),
		kgo.ConsumeTopics(*topic),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	for i := range 5 {
		r := &kgo.Record{
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("value-%d", i)),
			Headers: []kgo.RecordHeader{
				{Key: "source", Value: []byte("record-formatter")},
				{Key: "index", Value: []byte(fmt.Sprintf("%d", i))},
			},
			Timestamp: time.Now(),
		}
		cl.Produce(context.Background(), r, func(r *kgo.Record, err error) {
			if err != nil {
				log.Printf("error produce record: %v", err)
			}
		})
	}
	if err := cl.Flush(context.Background()); err != nil {
		log.Fatalf("unable to flush records: %v", err)
	}

	for {
		fs := cl.PollFetches(context.Background())
		if fs.IsClientClosed() {
			return
		}
		fs.EachRecord(func(r *kgo.Record) {
			out := formatter.AppendRecord(nil, r)
			os.Stdout.Write(out)
		})
		if fs.NumRecords() > 0 {
			break
		}
	}

	// You can also use the RecordReader to parse records from formatted
	// text. Here we create a reader that parses "key\tvalue\n" lines.
	inp := strings.NewReader("hello\tworld\ngoodbye\tplanet\n")
	reader, err := kgo.NewRecordReader(inp, "%k\t%v\n")
	if err != nil {
		log.Fatalf("unable to create record reader: %v", err)
	}
	for {
		r, err := reader.ReadRecord()
		if err != nil {
			break
		}
		r.Topic = *topic
		log.Printf("read record from input: key=%s value=%s", string(r.Key), string(r.Value))
	}
}
