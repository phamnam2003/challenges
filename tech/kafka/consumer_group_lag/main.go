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

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers  = flag.String("brokers", "localhost:9092", "comma delimited brokers to consume from")
	groups   = flag.String("group", "consumer_group_lag_group", "consumer group id")
	interval = flag.Duration("interval", 5*time.Second, "interval to print lag at")
)

func main() {
	flag.Parse()
	if len(*groups) == 0 {
		log.Fatal("missing require group flag")
	}
	groupList := strings.Split(*groups, ",")

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	adm := kadm.NewClient(cl)
	defer adm.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	printLag(ctx, adm, groupList)

	for {
		select {
		case <-ticker.C:
			printLag(ctx, adm, groupList)
		case <-ctx.Done():
			return
		}
	}
}

func printLag(ctx context.Context, adm *kadm.Client, groups []string) {
	lags, err := adm.Lag(ctx, groups...)
	if err != nil {
		log.Printf("error fetching consumer group lag: %v", err)
		return
	}

	lags.Each(func(l kadm.DescribedGroupLag) {
		if err := l.Error(); err != nil {
			log.Printf("group %s: error: %v", l.Group, err)
			return
		}

		log.Printf("group %s (state = %s, protocol = %s, total lag = %d)", l.Group, l.State, l.Protocol, l.Lag.Total())
		for _, tl := range l.Lag.TotalByTopic().Sorted() {
			log.Printf("topic %s: lag = %d", tl.Topic, tl.Lag)
		}

		for _, ml := range l.Lag.Sorted() {
			memberID := "(unassigned)"
			if ml.Member != nil {
				memberID = ml.Member.MemberID
			}

			if ml.Err != nil {
				log.Printf("member %s: topic %s partition %d: error: %v", memberID, ml.Topic, ml.Partition, ml.Err)
			} else {
				log.Printf("topic %s partition %d have %d lag, committed = %d, end = %d, member %s",
					ml.Topic, ml.Partition, ml.Lag, ml.Commit.At, ml.End.Offset, memberID,
				)
			}
		}
	})
}
