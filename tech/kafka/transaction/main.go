// This example demonstrates Kafka transactions. Run with -mode to select:
//
//   - produce: standalone transactional producer. Records are produced in
//     batches that are atomically committed or aborted. This mode alternates
//     between committing and aborting to demonstrate both paths.
//   - eos: exactly-once consume-transform-produce pipeline. Records are
//     consumed from one topic, transformed, and produced to another topic
//     with consumer offsets committed atomically in the same transaction.
//     This uses NewGroupTransactSession for the EOS session management.
//
// Both modes require -produce-to. The eos mode additionally requires -eos-to.
package main

import (
	"context"
	"flag"
	"log"
	"strconv"
	"strings"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers      = flag.String("brokers", "localhost:9092", "Comma-separated list of Kafka brokers to connect to")
	mode         = flag.String("mode", "produce", "Mode to run transaction: produce or eos")
	produceTo    = flag.String("produce-to", "", "topic to produce to or consume from (required)")
	group        = flag.String("group", "transaction-group", "consumer group to use in eos mode")
	eosTo        = flag.String("eos-to", "", "topic to produce to in eos mode (required for eos mode)")
	produceTrxID = flag.String("produce-trx-id", "eos-example-input-producer", "transactional ID to use for producer mode")
	consumeTrxID = flag.String("consume-trx-id", "eos-example-consume-eos-produce", "transactional ID to use for the eos consumer/producer")
)

func main() {
	flag.Parse()
	if len(*produceTo) == 0 {
		log.Fatal("missing topic to consume")
	}
	switch *mode {
	case "produce":
		inputProducer()
	case "eos":
		if *eosTo == "" {
			log.Fatal("missing topic to produce to in eos mode")
		}
		// The EOS pipeline needs input data, so run the input
		// producer concurrently.
		go inputProducer()
		eosConsumer()
	default:
		log.Fatalf("unknown mode: %s", *mode)
	}
}

// inputProducer demonstrates standalone transactional producing. Each
// iteration produces a batch of 10 records, then commits or aborts the
// transaction (alternating to show both paths).
func inputProducer() {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.DefaultProduceTopic(*produceTo),
		kgo.TransactionalID(*produceTrxID),
		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, func() string {
			return "[kgo-producer-logger]"
		})),
	)
	if err != nil {
		log.Fatalf("unable to create producer: %v", err)
	}

	ctx := context.Background()
	for doCommit := true; ; doCommit = !doCommit {
		if err := cl.BeginTransaction(); err != nil {
			log.Fatalf("unable to begin transaction: %v", err)
		}

		msg := "commit "
		if !doCommit {
			msg = "abort "
		}

		e := kgo.AbortingFirstErrPromise(cl)
		for i := range 10 {
			cl.Produce(ctx, kgo.StringRecord(msg+strconv.Itoa(i)), e.Promise())
			// Always evaluate e.Err() to avoid short-circuit issues
			// (e.g. doCommit && perr==nil would skip Err() if !doCommit).
			perr := e.Err()
			commit := kgo.TransactionEndTry(doCommit && perr == nil)

			switch err := cl.EndTransaction(ctx, commit); err {
			case nil:
				if doCommit {
					log.Println("transaction committed")
				} else {
					log.Println("transaction aborted")
				}
			case kerr.OperationNotAttempted:
				if err := cl.EndTransaction(ctx, kgo.TryAbort); err != nil {
					log.Fatalf("unable to abort transaction after failed commit: %v", err)
				}
			default:
				log.Fatalf("unable to end transaction: %v", err)
			}
		}
	}
}

// eosConsumer demonstrates exactly-once consume-transform-produce using
// NewGroupTransactSession. Records are consumed from -produce-to,
// transformed, and produced to -eos-to with consumer offsets committed
// atomically in the same transaction.
func eosConsumer() {
	sess, err := kgo.NewGroupTransactSession(
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		kgo.DefaultProduceTopic(*eosTo),
		kgo.TransactionalID(*consumeTrxID),
		kgo.FetchIsolationLevel(kgo.ReadCommitted()),
		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, func() string {
			return "[eos consumer]"
		})),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*produceTo),
		kgo.RequireStableFetchOffsets(),
	)
	if err != nil {
		log.Fatalf("unable to create EOS session: %v", err)
	}
	defer sess.Close()

	ctx := context.Background()

	for {
		fetches := sess.PollFetches(ctx)

		if fetchErrs := fetches.Errors(); len(fetchErrs) > 0 {
			for _, fetchErr := range fetchErrs {
				log.Printf("fetch error: topic=%s partition=%d error=%v", fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			}

			// The errors may be fatal for the partition (auth
			// problems), but we can still process any records if
			// there are any.
		}

		if err := sess.Begin(); err != nil {
			log.Fatalf("unable to start transactions: %v", err)
		}

		e := kgo.AbortingFirstErrPromise(sess.Client())
		fetches.EachRecord(func(r *kgo.Record) {
			sess.Produce(ctx, kgo.StringRecord("eos "+string(r.Value)), e.Promise())
		})

		committed, err := sess.End(ctx, e.Err() == nil)
		if committed {
			log.Printf("eos committed successful")
		} else {
			// A failed End always means an error occurred, because
			// End retries as appropriate.
			log.Fatalf("unable to eos commit: %v", err)
		}
	}
}
