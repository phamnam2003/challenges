package main

import (
	"context"
	"flag"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
)

var (
	brokers             = flag.String("brokers", "localhost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	topic               = flag.String("topic", "bench-sarama-topic", "Kafka topic to produce to")
	recordBytes         = flag.Int("record-bytes", 100, "bytes per record (producing)")
	noCompression       = flag.Bool("no-compression", false, "disable snappy compression (producing)")
	consume             = flag.Bool("consume", false, "consume messages instead of producing")
	group               = flag.String("group", "bench-sarama-group", "if non-empty, group to use for consuming rather than direct partition consuming (consuming)")
	dialTLS             = flag.Bool("tls", false, "if true, use tls for connecting")
	saslMethod          = flag.String("sasl-method", "", "if non-empty, sasl method to use (must specify all options; supports plain, scram-sha-256, scram-sha-512)")
	saslUser            = flag.String("sasl-user", "", "if non-empty, username to use for sasl (must specify all options)")
	saslPass            = flag.String("sasl-pass", "", "if non-empty, password to use for sasl (must specify all options)")
	rateBytes, rateRecs int64
)

func printRate() {
	for range time.Tick(time.Second) {
		recs := atomic.SwapInt64(&rateRecs, 0)
		bytes := atomic.SwapInt64(&rateBytes, 0)
		log.Printf("%0.2f MiB/s; %0.2fk records/s\n", float64(bytes)/(1024*1024), float64(recs)/1000)
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

type consumer struct{}

func (*consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (*consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		atomic.AddInt64(&rateRecs, 1)
		atomic.AddInt64(&rateBytes, int64(len(msg.Value)))
		sess.MarkMessage(msg, "")
	}
	return nil
}

func main() {
	flag.Parse()

	go printRate()

	conf := sarama.NewConfig()
	conf.Version = sarama.V2_8_0_0

	if *dialTLS {
		conf.Net.TLS.Enable = true
	}

	if *saslMethod != "" || *saslUser != "" || *saslPass != "" {
		if *saslMethod == "" || *saslUser == "" || *saslPass == "" {
			log.Fatal("all of -sasl-method, -sasl-user, -sasl-pass must be specified if any are")
		}
		method := strings.ToLower(*saslMethod)
		method = strings.ReplaceAll(method, "-", "")
		method = strings.ReplaceAll(method, "_", "")

		conf.Net.SASL.Enable = true
		conf.Net.SASL.User = *saslUser
		conf.Net.SASL.Password = *saslPass
		conf.Net.SASL.Version = 1

		switch method {
		case "plain":
			conf.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		case "scramsha256":
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		case "scramsha512":
			conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		default:
			log.Fatalf("unsupported sasl method: %s", *saslMethod)
		}
	}

	switch *consume {
	case false:
		conf.Producer.RequiredAcks = sarama.WaitForAll
		conf.Producer.Compression = sarama.CompressionSnappy
		conf.Producer.Return.Successes = true

		p, err := sarama.NewAsyncProducer(strings.Split(*brokers, ","), conf)
		if err != nil {
			log.Fatalf("failed to create producer: %s", err)
		}
		go func() {
			for err := range p.Errors() {
				if err != nil {
					log.Fatalf("failed to produce message: %s", err)
				}
			}
		}()
		go func() {
			for range p.Successes() {
				atomic.AddInt64(&rateBytes, int64(*recordBytes))
				atomic.AddInt64(&rateRecs, 1)
			}
		}()

		var num int64
		for {
			p.Input() <- &sarama.ProducerMessage{
				Topic: *topic,
				Value: sarama.ByteEncoder(newValue(num)),
			}
			num++
		}

	case true:
		conf.Consumer.Return.Errors = true
		conf.Consumer.Offsets.Initial = sarama.OffsetOldest

		if *group != "" {
			c, err := sarama.NewConsumer(strings.Split(*brokers, ","), conf)
			if err != nil {
				log.Fatalf("unable create consumer: %v", err)
			}
			ps, err := c.Partitions(*topic)
			if err != nil {
				log.Fatalf("unable to get partitions: %v", err)
			}

			for _, p := range ps {
				go func(partition int32) {
					pc, err := c.ConsumePartition(*topic, partition, sarama.OffsetOldest)
					if err != nil {
						log.Fatalf("unable to consume partition %d: %v", partition, err)
					}
					for {
						msg := <-pc.Messages()
						atomic.AddInt64(&rateBytes, int64(len(msg.Value)))
						atomic.AddInt64(&rateRecs, 1)
					}
				}(p)
			}
			select {}
		}

		g, err := sarama.NewConsumerGroup(strings.Split(*brokers, ","), *group, conf)
		if err != nil {
			log.Fatalf("failed to create consumer group: %s", err)
		}
		go func() {
			for err := range g.Errors() {
				if err != nil {
					log.Fatalf("error from consumer group: %s", err)
				}
			}
		}()

		for {
			err := g.Consume(context.Background(), []string{*topic}, new(consumer))
			if err != nil {
				log.Fatalf("consumer group err: %v", err)
			}
		}
	}
}
