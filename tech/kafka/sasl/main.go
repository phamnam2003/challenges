package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

var (
	brokers  = flag.String("brokers", "localhost:9092", "comma delimited list of seed brokers")
	method   = flag.String("method", "plain", "sasl mechanism to use (plain, scram-sha-512, scram-sha-256, aws-msk-iam)")
	username = flag.String("username", "", "sasl username")
	password = flag.String("password", "", "sasl password")
)

func main() {
	flag.Parse()
	if *method != "aws-msk-iam" && (*username == "" || *password == "") {
		log.Fatal("username and password not provided yet")
	}

	// tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}}
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(*brokers, ",")...),
		// kgo.Dialer(tlsDialer.DialContext),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
	}

	switch *method {
	case "plain":
		opts = append(opts, kgo.SASL(plain.Auth{
			User: *username,
			Pass: *password,
		}.AsMechanism()))

	case "scram-sha-256":
		opts = append(opts, kgo.SASL(scram.Auth{
			User: *username,
			Pass: *password,
		}.AsSha256Mechanism()))

	case "scram-sha-512":
		opts = append(opts, kgo.SASL(scram.Auth{
			User: *username,
			Pass: *password,
		}.AsSha512Mechanism()))

	case "aws-msk-iam":
		ctx := context.Background()
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("unable to load AWS config: %v", err)
		}

		opts = append(opts, kgo.SASL(aws.ManagedStreamingIAM(func(ctx context.Context) (aws.Auth, error) {
			creds, err := cfg.Credentials.Retrieve(ctx)
			if err != nil {
				return aws.Auth{}, err
			}
			return aws.Auth{
				AccessKey:    creds.AccessKeyID,
				SecretKey:    creds.SecretAccessKey,
				SessionToken: creds.SessionToken,
				UserAgent:    "franz-go-kafka/plain-auth",
			}, nil
		})))
	default:
		log.Fatalf("not support sasl method %s", *method)
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatalf("unable to create client: %v", err)
	}
	defer cl.Close()

	if err = cl.Ping(context.Background()); err != nil {
		log.Fatalf("unable to ping cluster: %v", err)
	}
	log.Printf("connected to kafka cluster with sals mechanism %s with [username: %s, password: %s]", *method, *username, *password)
}
