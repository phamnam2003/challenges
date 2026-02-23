// This example demonstrates runtime consumer control: dynamically adding and
// removing topics/partitions, and pausing/resuming fetch operations.
//
// These APIs are useful for:
//   - Dynamically discovering and consuming new topics
//   - Implementing backpressure or priority-based consuming
//   - Temporarily halting consumption from specific topics/partitions
//
// Run with -mode=add-topics, -mode=add-partitions, or -mode=pause to see
// each feature in action.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers = flag.String("brokers", "localhost:9092", "comma delimited list of seed brokers")
	mode    = flag.String("mode", "add-topics", "control mode: add-topics, add-partitions, or pause")
)

func main() {
	flag.Parse()

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.ConsumeTopics("topic-a"),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		log.Fatalf("unable to create kgo client: %v", err)
	}
	defer cl.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	switch *mode {
	case "add-topics":
		demoAddTopics(ctx, cl)
	case "add-partitions":
		demoAddPartitions(ctx, cl)
	case "pause":
		demoPause(ctx, cl)
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

// demoAddTopics shows AddConsumeTopics: after a delay, "topic-b" is added
// to the set of consumed topics at runtime.
func demoAddTopics(ctx context.Context, cl *kgo.Client) {
	go func() {
		time.Sleep(3 * time.Second)
		cl.AddConsumeTopics("topic-b")
		log.Printf("added topic-b, now consuming: %v", cl.GetConsumeTopics())
	}()
	pollLoop(ctx, cl)
}

// demoAddPartitions shows AddConsumePartitions and RemoveConsumePartitions:
// specific partitions of "topic-c" are added, then one is removed.
func demoAddPartitions(ctx context.Context, cl *kgo.Client) {
	go func() {
		time.Sleep(3 * time.Second)
		log.Printf("adding topic-c partitions 0 and 1")
		cl.AddConsumePartitions(map[string]map[int32]kgo.Offset{
			"topic-c": {
				0: kgo.NewOffset().AtStart(),
				1: kgo.NewOffset().AtEnd(),
			},
		})

		log.Printf("now consuming: %v", cl.GetConsumeTopics())

		time.Sleep(3 * time.Second)
		cl.RemoveConsumePartitions(map[string][]int32{
			"topic-c": {1},
		})
		log.Printf("removed topic-c partition 1, now consuming: %v", cl.GetConsumeTopics())
	}()
	pollLoop(ctx, cl)
}

// demoPause shows PauseFetchTopics and ResumeFetchTopics: topic-a is paused
// after some records, then resumed after a cooldown.
func demoPause(ctx context.Context, cl *kgo.Client) {
	var (
		count  = 0
		paused = false
	)

	for {
		fetches := cl.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Fatalf("error consuming topic %s partition %d: %v", topic, partition, err)
		})

		fetches.EachRecord(func(r *kgo.Record) {
			log.Printf("topic=%s partition=%d offset=%d value=%s", r.Topic, r.Partition, r.Offset, string(r.Value))
			count++
		})

		if !paused && count >= 50 {
			cl.PauseFetchTopics("topic-a")
			paused = true
			log.Printf("paused topic-a after %d records", count)

			// Query currently paused topics (no args = query only).
			log.Printf("currently paused: %v", cl.PauseFetchTopics())

			go func() {
				time.Sleep(3 * time.Second)
				cl.ResumeFetchTopics("topic-a")
				log.Println("resumed topic-a")
			}()
		}
	}
}

func pollLoop(ctx context.Context, cl *kgo.Client) {
	for {
		fetches := cl.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Printf("error consuming topic %s partition %d: %v", topic, partition, err)
		})

		fetches.EachRecord(func(r *kgo.Record) {
			log.Printf("topic=%s partition=%d offset=%d value=%s", r.Topic, r.Partition, r.Offset, string(r.Value))
		})
	}
}
