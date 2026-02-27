package main

import (
	"flag"

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

func main() {
	flag.Parse()

	rcl, er := sr.NewClient(sr.URLs(*registry))
}
