package main

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestProduceMultipleRecordsWithVirtualNetwork(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var stack kfake.VirtualNetwork

		cluster, err := kfake.NewCluster(
			kfake.NumBrokers(1),
			kfake.Ports(9093),
			kfake.SeedTopics(3, "multi-topic"),
			kfake.ListenFn(stack.Listen),
		)
		if err != nil {
			t.Fatalf("unable to create cluster: %v", err)
		}
		defer cluster.Close()

		client, err := kgo.NewClient(
			kgo.SeedBrokers(cluster.ListenAddrs()...),
			kgo.DefaultProduceTopic("multi-topic"),
			kgo.ConsumeTopics("multi-topic"),
			kgo.Dialer(stack.DialContext),
		)
		if err != nil {
			t.Fatalf("unable to create client: %v", err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		const numRecs = 10
		for i := range numRecs {
			record := &kgo.Record{Value: []byte{byte(i)}}
			if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
				t.Fatalf("unexpected produce error: %v", err)
			}
		}

		var consumed int
		for consumed < numRecs {
			fetches := client.PollFetches(ctx)
			consumed += fetches.NumRecords()
		}
		if consumed != numRecs {
			t.Errorf("expected %d records, got %d", numRecs, consumed)
		}
	})
}

func TestTimeoutWithVirtualNetwork(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var stack kfake.VirtualNetwork

		cluster, err := kfake.NewCluster(
			kfake.NumBrokers(1),
			kfake.Ports(9093),
			kfake.SeedTopics(3, "timeout-topic"),
			kfake.ListenFn(stack.Listen),
		)
		if err != nil {
			t.Fatalf("unable to create cluster: %v", err)
		}
		defer cluster.Close()

		consumer, err := kgo.NewClient(
			kgo.SeedBrokers(cluster.ListenAddrs()...),
			kgo.ConsumeTopics("timeout-topic"),
			kgo.Dialer(stack.DialContext),
			kgo.FetchMaxWait(100*time.Millisecond),
		)
		if err != nil {
			t.Fatalf("unable to create consumer: %v", err)
		}
		defer consumer.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		fetches := consumer.PollFetches(ctx)
		if fetches.NumRecords() != 0 {
			t.Errorf("expected 0 records from empty topic, got %d", fetches.NumRecords())
		}
	})
}
