package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

var (
	brokers  = flag.String("brokers", "localhost:8081", "comma delimited list of seed brokers")
	topic    = flag.String("topic", "schema-registry-topic", "topic to produce to")
	registry = flag.String("registry", "localhost:8081", "schema registry port to talk to")

	schemaText = `{
		"type": "record",
		"name": "simple",
		"namespace": "org.hamba.avro",
		"fields" : [
			{"name": "a", "type": "long"},
			{"name": "b", "type": "string"}
		]
	}`
)

type example struct {
	A int64  `arvo:"a"`
	B string `arvo:"b"`
}

func main() {
	flag.Parse()

	rcl, err := sr.NewClient(sr.URLs(*registry))
	if err != nil {
		log.Fatalf("unable create schema registry client: %v", err)
	}

	ss, err := rcl.CreateSchema(context.Background(), *topic+"-value", sr.Schema{
		Schema: schemaText,
		Type:   sr.TypeAvro,
	})
	if err != nil {
		log.Fatalf("unable create arvo schema: %v", err)
	}
	log.Printf("created or reusing schema subject %q version %d id %d", ss.Subject, ss.Version, ss.ID)

	// Setup our serializer / deserializer.
	avroSchema, err := avro.Parse(schemaText)
	if err != nil {
		log.Fatalf("unable to parse avro schema: %v", err)
	}

	var serde sr.Serde
	serde.Register(
		ss.ID,
		example{},
		sr.EncodeFn(func(a any) ([]byte, error) {
			return avro.Marshal(avroSchema, a)
		}),
		sr.DecodeFn(func(b []byte, v any) error {
			return avro.Unmarshal(avroSchema, b, v)
		}),
	)

	// Loop producing & consuming.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.DefaultProduceTopic(*topic),
		kgo.ConsumeTopics(*topic),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		log.Fatalf("unable to create kafka client: %v", err)
	}
	defer cl.Close()

	for {
		cl.Produce(
			context.Background(),
			&kgo.Record{
				Value: serde.MustEncode(example{
					A: time.Now().Unix(),
					B: "schema registry with avro",
				}),
			},
			func(r *kgo.Record, err error) {
				if err != nil {
					log.Fatalf("unable to produce record kafka: %v", err)
				}

				log.Printf("produced simple record, values bytes: %x", r.Value)
			},
		)

		fs := cl.PollFetches(context.Background())
		if fs.IsClientClosed() {
			return
		}
		fs.EachRecord(func(r *kgo.Record) {
			var ex example
			err := serde.Decode(r.Value, &ex)
			if err != nil {
				log.Fatalf("cannot decode record value: %v", err)
			}
			log.Printf("consumed record, value: %+v", ex)
		})

		time.Sleep(1 * time.Second)
	}
}
