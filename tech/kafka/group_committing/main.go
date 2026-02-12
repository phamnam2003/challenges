package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	seedBrokers = flag.String("brokers", "localhost:9092", "comma delimited list of seed brokers")
	topicName   = flag.String("topic", "group-committing-topic", "topic to produce to")
	style       = flag.String("style", "autocommit", "committing style: autocommit|records|uncommitted|marks|kadm")
	group       = flag.String("group", "group-committing-group", "group to comsume within")
	logger      = flag.Bool("logger", false, "if true, enable an info level logger")
)

func main() {
	flag.Parse()

	var styleNum = 0
	switch {
	case strings.HasPrefix("autocommit", *style):
	case strings.HasPrefix("records", *style):
		styleNum = 1
	case strings.HasPrefix("uncommitted", *style):
		styleNum = 2
	case strings.HasPrefix("marks", *style):
		styleNum = 3
	case strings.HasPrefix("kadm", *style):
		styleNum = 4
	default:
		log.Fatalf("unknown style %q", *style)
	}

	seeds := kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...)
	// The kadm style bypasses the consumer group protocol entirely.
	// It fetches committed offsets via kadm, consumes via direct
	// partition assignment, and commits offsets back via kadm.
	if styleNum == 4 {
		consumeKadm(seeds)
		return
	}

	opts := []kgo.Opt{
		seeds,
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topicName),
	}
	switch styleNum {
	case 1, 2:
		opts = append(opts, kgo.DisableAutoCommit())
	case 3:
		// AutoCommitMarks causes autocommitting to only commit
		// offsets that have been explicitly marked, rather than
		// committing all polled offsets. BlockRebalanceOnPoll
		// ensures no rebalance happens between polling and marking.
		opts = append(opts,
			kgo.AutoCommitMarks(),
			kgo.BlockRebalanceOnPoll(),
			kgo.OnPartitionsRevoked(func(ctx context.Context, c *kgo.Client, m map[string][]int32) {
				if err := c.CommitMarkedOffsets(ctx); err != nil {
					log.Printf("revoke commit failed: %v", err)
				}
			}),
		)
	}

	if *logger {
		opts = append(opts, kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)))
	}
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()
	go consume(cl, styleNum)

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)

	<-sigs
	log.Println("received interrupt signal; closing client")
	done := make(chan struct{})
	go func() {
		defer close(done)
		cl.Close()
	}()

	select {
	case <-sigs:
		log.Println("received second interrupt signal; quitting without waiting for graceful close")
	case <-done:
	}
}

// consumeKadm demonstrates consuming without the consumer group protocol.
// It uses kadm to fetch previously committed offsets, then consumes from
// those offsets using direct partition assignment. After processing, it
// commits offsets back via kadm.
func consumeKadm(seeds kgo.Opt) {
	adm, err := kgo.NewClient(seeds)
	if err != nil {
		log.Fatalf("unable to create kadm client: %v", err)
	}
	admCli := kadm.NewClient(adm)

	// FetchOffsetsForTopics returns committed offsets for the group,
	// defaulting to -1 for partitions that have no prior commit.
	os, err := admCli.FetchOffsetsForTopics(context.Background(), *group, *topicName)
	if err != nil {
		log.Fatalf("unable to fetch offsets for group %q: %v", *group, err)
	}
	cl, err := kgo.NewClient(seeds, kgo.ConsumePartitions(os.KOffsets()))
	if err != nil {
		log.Fatalf("unable to create consuming client: %v", err)
	}
	defer cl.Close()

	log.Println("waiting for one record...")
	fs := cl.PollRecords(context.Background(), 1)
	if err := admCli.CommitAllOffsets(context.Background(), *group, kadm.OffsetsFromFetches(fs)); err != nil {
		log.Fatalf("unable to commit offsets for group %q: %v", *group, err)
	}
	r := fs.Records()[0]
	log.Printf("commited record on partition %d at offset %d", r.Partition, r.Offset)
}

func consume(cl *kgo.Client, style int) {
	for {
		fetches := cl.PollFetches(context.Background())
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			log.Fatalf("fetch error on topic %s, partition %d: %v", t, p, err)
		})

		switch style {
		case 0:
			var seen int
			fetches.EachRecord(func(r *kgo.Record) {
				seen++
			})
			log.Printf("processed %d records -- autocommitting now allows the **prior** poll to be available for committing, nothing can be lost!", seen)

		case 1:
			var rs []*kgo.Record
			fetches.EachRecord(func(r *kgo.Record) {
				rs = append(rs, r)
			})
			if err := cl.CommitRecords(context.Background(), rs...); err != nil {
				log.Printf("commit record failed: %v", err)
				continue
			}
			log.Printf("committed %d records individually--this demo does this in a naive way by just hanging on to all records, but you could just hang on to the max offset record per topic/partition!", len(rs))

		case 2:
			var seen int
			fetches.EachRecord(func(*kgo.Record) {
				seen++
			})
			if err := cl.CommitUncommittedOffsets(context.Background()); err != nil {
				log.Printf("commit records failed: %v", err)
				continue
			}
			log.Printf("committed %d records successfully--the recommended pattern, as followed in this demo, is to commit all uncommitted offsets after each poll!\n", seen)

		case 3:
			// Build the full mark map across all partitions, then
			// mark everything at once. Only marked offsets are
			// autocommitted.
			marks := make(map[string]map[int32]kgo.EpochOffset)
			fetches.EachPartition(func(ftp kgo.FetchTopicPartition) {
				if len(ftp.Records) == 0 {
					return
				}
				last := ftp.Records[len(ftp.Records)-1]
				if marks[ftp.Topic] == nil {
					marks[ftp.Topic] = make(map[int32]kgo.EpochOffset)
				}
				marks[ftp.Topic][ftp.Partition] = kgo.EpochOffset{
					Offset: last.Offset + 1,
					Epoch:  last.LeaderEpoch,
				}
				log.Printf("marked %s p%d through offset %d\n", ftp.Topic, ftp.Partition, last.Offset)
			})

			cl.MarkCommitOffsets(marks)
			// Allow a blocked rebalance to proceed now that we
			// have marked everything from this poll.
			cl.AllowRebalance()
		}
	}
}
