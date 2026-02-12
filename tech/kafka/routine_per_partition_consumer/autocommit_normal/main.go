package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedBrokers = flag.String("brokers", "localhost:9092", "comma delimited brokers to consume from")
	topic       = flag.String("topic", "autocommit_normal", "topic to consume from")
	group       = flag.String("group", "autocommit_group", "consumer group id")

	globalRecs  int64
	globalBytes int64
)

type pconsumer struct {
	quit chan struct{}
	recs chan []*kgo.Record
}

type splitConsume struct {
	mu        sync.Mutex
	consumers map[string]map[int32]pconsumer
}

func (pc pconsumer) consume(topic string, partition int32) {
	log.Printf("starting, t %s p %d", topic, partition)
	defer log.Printf("killing, t %s p %d", topic, partition)

	var (
		nrecs  int
		nbytes int
		ticket = time.NewTicker(time.Second)
	)
	defer ticket.Stop()

	for {
		select {
		case <-pc.quit:
			return

		case recs := <-pc.recs:
			nrecs += len(recs)
			atomic.AddInt64(&globalRecs, int64(len(recs)))
			for _, rec := range recs {
				nbytes += len(rec.Value)
				atomic.AddInt64(&globalBytes, int64(len(rec.Value)))
			}

		case t := <-ticket.C:
			log.Printf("[%s] t %s p %d consumed %0.2f MiB/s, %0.2fk records/s",
				t.Format("15:04:05.999"),
				topic,
				partition,
				float64(nbytes)/(1024*1024),
				float64(nrecs)/1000,
			)
			nrecs, nbytes = 0, 0
		}
	}
}

func (s *splitConsume) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for topic, partitions := range assigned {
		if s.consumers[topic] == nil {
			s.consumers[topic] = make(map[int32]pconsumer)
		}

		for _, partition := range partitions {
			pc := pconsumer{
				quit: make(chan struct{}),
				recs: make(chan []*kgo.Record, 10),
			}
			s.consumers[topic][partition] = pc
			go pc.consume(topic, partition)
		}
	}
}

func (s *splitConsume) lost(_ context.Context, cl *kgo.Client, lost map[string][]int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for topic, partitions := range lost {
		ptopics := s.consumers[topic]
		for _, partition := range partitions {
			pc := ptopics[partition]
			delete(ptopics, partition)
			if len(ptopics) == 0 {
				delete(s.consumers, topic)
			}
			close(pc.quit)
		}
	}
}

func (s *splitConsume) poll(cl *kgo.Client) {
	for {
		fetches := cl.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Printf("fetch error, t %s p %d: %v", topic, partition, err)
			panic(err)
		})
		fetches.EachTopic(func(ft kgo.FetchTopic) {
			s.mu.Lock()
			tconsumers := s.consumers[ft.Topic]
			s.mu.Unlock()
			if tconsumers == nil {
				return
			}
			ft.EachPartition(func(fp kgo.FetchPartition) {
				pc, ok := tconsumers[fp.Partition]
				if !ok {
					return
				}

				select {
				case pc.recs <- fp.Records:
				case <-pc.quit:
				}
			})
		})
	}
}

func main() {
	flag.Parse()

	go func() {
		for t := range time.Tick(time.Second) {
			log.Printf("[%s] globally consumed %0.2f MiB/s, %0.2fk records/s",
				t.Format(time.TimeOnly),
				float64(atomic.SwapInt64(&globalBytes, 0))/(1024*1024),
				float64(atomic.SwapInt64(&globalRecs, 0))/1000,
			)
		}
	}()

	s := &splitConsume{
		consumers: make(map[string]map[int32]pconsumer),
	}
	if len(*group) == 0 {
		log.Fatal("missing required group")
	}
	if len(*topic) == 0 {
		log.Fatal("missing required topic")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.OnPartitionsAssigned(s.assigned),
		kgo.OnPartitionsRevoked(s.lost),
		kgo.OnPartitionsLost(s.lost),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	}
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		fmt.Println("received interrupt signal; closing client")
		cl.Close()
		<-sigs
		fmt.Println("received second interrupt; exiting")
		os.Exit(1)
	}()
	s.poll(cl)
}
