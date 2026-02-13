package main

import (
	"context"
	"flag"
	"fmt"
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
	seedBrokers = flag.String("brokers", "localhost:9092", "list brokers separate by comma")
	topic       = flag.String("topic", "manual_commit_goroutine_per_partition_topic", "topic to consume")
	group       = flag.String("group", "manual_commit_goroutine_per_partition_group", "group to consume in")
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
	recs chan []*kgo.Record
}

func (pc *pconsumer) consume() {
	defer close(pc.done)

	log.Printf("starting consume for t %s p %d", pc.topic, pc.partition)
	defer log.Printf("closing consume for t %s p %d", pc.topic, pc.partition)
	for {
		select {
		case <-pc.quit:
			return

		case recs := <-pc.recs:
			time.Sleep(time.Duration(rand.IntN(150)+100) * time.Millisecond) // present process
			log.Printf("some sort of work done, about to commit t %s p %d", pc.topic, pc.partition)
			err := pc.cl.CommitRecords(context.Background(), recs...)
			if err != nil {
				log.Printf("error when committing offsets to kafka err: %v t: %s p: %d offset: %d",
					err,
					pc.topic,
					pc.partition,
					recs[len(recs)-1].Offset+1,
				)
			}
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
		for _, partition := range partitions {
			pc := &pconsumer{
				cl:        cl,
				topic:     topic,
				partition: partition,
				quit:      make(chan struct{}),
				done:      make(chan struct{}),
				recs:      make(chan []*kgo.Record),
			}

			s.consumers[tp{topic, partition}] = pc
			go pc.consume()
		}
	}
}

// In this example, each partition consumer commits itself. Those commits will
// fail if partitions are lost, but will succeed if partitions are revoked. We
// only need one revoked or lost function (and we name it "lost").
func (s *splitConsume) lost(_ context.Context, cl *kgo.Client, lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, partition := range partitions {
			tp := tp{topic, partition}
			pc := s.consumers[tp]
			delete(s.consumers, tp)
			close(pc.quit)
			log.Printf("waiting for work to finish t %s p %d", topic, partition)
			wg.Add(1)
			go func() {
				<-pc.done
				wg.Done()
			}()
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
		fetches := cl.PollRecords(context.Background(), 10_000)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(s string, i int32, err error) {
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
			s.consumers[tp].recs <- ftp.Records
		})

		cl.AllowRebalance()
	}
}

func main() {
	flag.Parse()

	if len(*group) == 0 {
		log.Fatal("no group to consume topic")
	}
	if len(*topic) == 0 {
		log.Fatal("no topic to consume")
	}

	s := &splitConsume{
		consumers: make(map[tp]*pconsumer),
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...),
		kgo.AllowAutoTopicCreation(),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.OnPartitionsAssigned(s.assigned),
		kgo.OnPartitionsRevoked(s.lost),
		kgo.OnPartitionsLost(s.lost),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	}
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable create client: %v", err)
	}
	defer cl.Close()

	if err = cl.Ping(context.Background()); err != nil { // check connectivity to cluster
		log.Fatalf("cannot ping to kafka: %v", err)
	}

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
