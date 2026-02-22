package main

import (
	"context"
	"flag"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers = flag.String("brokers", "localhost:9092", "comma delimited brokers to consume from")
	topic   = flag.String("topic", "autocommit_marks", "topic to consume from")
	group   = flag.String("group", "autocommit_marks_group", "consumer group id")
)

type tp struct {
	t string
	p int32
}

type pconsumer struct {
	cl        *kgo.Client
	topic     string
	partition int32

	quit chan struct{}
	done chan struct{}
	recs chan kgo.FetchTopicPartition
}

func (pc *pconsumer) consume() {
	defer close(pc.done)
	log.Printf("starting consume for t %s p %d", pc.topic, pc.partition)
	defer log.Printf("closing consume for t %s p %d", pc.topic, pc.partition)

	for {
		select {
		case <-pc.quit:
			return
		case p := <-pc.recs:
			time.Sleep(time.Duration(rand.IntN(150)+100) * time.Millisecond) // simulate work
			log.Printf("Some sort of work done, about to commit t %s p %d\n", pc.topic, pc.partition)
			pc.cl.MarkCommitRecords(p.Records...)
		}
	}
}

type splitConsume struct {
	// Using BlockRebalanceOnPoll means we do not need a mu to manage
	// consumers, unlike the autocommit normal example.
	consumers map[tp]*pconsumer
}

func (s *splitConsume) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
	for topic, partitions := range assigned {
		for _, p := range partitions {
			pc := &pconsumer{
				cl:        cl,
				topic:     topic,
				partition: p,
				quit:      make(chan struct{}),
				done:      make(chan struct{}),
				recs:      make(chan kgo.FetchTopicPartition, 5),
			}
			s.consumers[tp{topic, p}] = pc
			go pc.consume()
		}
	}
}

func (s *splitConsume) revoked(ctx context.Context, cl *kgo.Client, revoked map[string][]int32) {
	s.killConsumers(revoked)
	if err := cl.CommitMarkedOffsets(ctx); err != nil {
		log.Printf("failed to commit offsets on revoke: %v", err)
	}
}

func (s *splitConsume) lost(_ context.Context, _ *kgo.Client, lost map[string][]int32) {
	s.killConsumers(lost)
	// Losing means we cannot commit: an error happened.
}

func (s *splitConsume) killConsumers(lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, p := range partitions {
			tp := tp{topic, p}
			if pc, ok := s.consumers[tp]; ok {
				delete(s.consumers, tp)
				close(pc.quit)
				log.Printf("waiting for work to finish t %s p %d", pc.topic, pc.partition)
				wg.Add(1)
				go func() {
					<-pc.done
					wg.Done()
				}()
			}
		}
	}
}

func (s *splitConsume) poll(cl *kgo.Client) {
	for {
		// PollRecords is strongly recommended when using
		// BlockRebalanceOnPoll. You can tune how many records to
		// process at once (upper bound -- could all be on one
		// partition), ensuring that your processor loops complete fast
		// enough to not block a rebalance too long.
		fetches := cl.PollRecords(context.Background(), 10000)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			// Note: you can delete this block, which will result
			// in these errors being sent to the partition
			// consumers, and then you can handle the errors there.
			panic(err)
		})

		fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
			tp := tp{ftp.Topic, ftp.Partition}
			// Since we are using BlockRebalanceOnPoll, we can be
			// sure this partition consumer exists:
			//
			// * onAssigned is guaranteed to be called before we
			// fetch offsets for newly added partitions
			//
			// * onRevoked waits for partition consumers to quit
			// and be deleted before re-allowing polling.
			s.consumers[tp].recs <- ftp
		})
		cl.AllowRebalance()
	}
}

func main() {
	flag.Parse()
	if len(*group) == 0 {
		log.Fatal("group is required")
	}
	if len(*topic) == 0 {
		log.Fatal("topic is required")
	}

	s := &splitConsume{
		consumers: make(map[tp]*pconsumer),
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.OnPartitionsAssigned(s.assigned),
		kgo.OnPartitionsRevoked(s.revoked),
		kgo.OnPartitionsLost(s.lost),
		kgo.AutoCommitMarks(),
		kgo.BlockRebalanceOnPoll(),
		kgo.AllowAutoTopicCreation(),
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
		log.Println("received interrupt signal; closing client")
		cl.Close()
		<-sigs
		log.Println("received second interrupt; exiting")
		os.Exit(1)
	}()

	s.poll(cl)
}
