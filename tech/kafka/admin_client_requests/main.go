package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/pkg/kversion"
)

var (
	seedBrokers = flag.String("brokers", "localhost:9092", "comma delimited list of seed brokers")
	topicName   = flag.String("topic", "admin-client-req-topic", "topic to create")
	mode        = flag.String("mode", "kadm", "admin mode: kadm, kmsg")

	counter     int32 = 0
	partition   int32 = 3
	replication int16 = 1
)

func demoKadm(ctx context.Context, cl *kgo.Client) {
	admin := kadm.NewClient(cl)

	topics, err := admin.ListTopics(ctx)
	if err != nil {
		log.Fatalf("unable get list topics from kadm: %v", err)
	}
	if topics.Has(*topicName) {
		log.Printf("topic %s already exists", *topicName)
		return
	}

	resp, err := admin.CreateTopic(ctx, partition, replication, nil, *topicName)
	if err != nil {
		log.Fatalf("unable create topic with %d partitions, %d replications: %v", partition, replication, err)
	}
	log.Printf("created topic %s with %d partitions, %d replications. %+v", *topicName, partition, replication, resp)

	counter++
	if counter <= 5 {
		demoKadm(ctx, cl)
	}
}
func demoKmsg(ctx context.Context, cl *kgo.Client) {
	// Create a topic.
	{
		req := kmsg.NewPtrCreateTopicsRequest()
		t := kmsg.NewCreateTopicsRequestTopic()
		t.Topic = *topicName
		t.NumPartitions = partition
		t.ReplicationFactor = replication
		req.Topics = append(req.Topics, t)

		res, err := req.RequestWith(ctx, cl)
		if err != nil {
			log.Fatalf("unable create topic with kmsg: %v", err)
		}
		if len(res.Topics) != 1 {
			log.Fatalf("unexpected number of topics in response: %d", len(res.Topics))
		}
		if err := kerr.ErrorForCode(res.Topics[0].ErrorCode); err != nil {
			log.Fatalf("error creating topic: %v", err)
		} else {
			log.Printf("created topic %s with %d partitions, %d replications", *topicName, partition, replication)
		}
	}

	// Issue a metadata request for the topic.
	{
		req := kmsg.NewPtrMetadataRequest()
		t := kmsg.NewMetadataRequestTopic()
		t.Topic = topicName
		req.Topics = append(req.Topics, t)

		res, err := req.RequestWith(ctx, cl)
		if err != nil {
			log.Fatalf("unable get metadata with kmsg: %v", err)
		}
		for _, topic := range res.Topics {
			if err := kerr.ErrorForCode(topic.ErrorCode); err != nil {
				log.Printf("topic %s: error %v", *topic.Topic, err)
				continue
			}
			log.Printf("topic %s has %d partitions", *topic.Topic, len(topic.Partitions))
		}
		log.Printf("cluster has %d brokers", len(res.Brokers))
	}
}

func main() {
	flag.Parse()

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*seedBrokers, ",")...),
		kgo.MaxVersions(kversion.V2_4_0()),
	)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	switch *mode {
	case "kadm":
		demoKadm(ctx, cl)
	case "kmsg":
		demoKmsg(ctx, cl)
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}
