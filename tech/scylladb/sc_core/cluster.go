// Package sccore serves as the core library for ScyllaDB-related operations.
package sccore

import (
	"time"

	"github.com/gocql/gocql"
)

func CreateCluster(consistency gocql.Consistency, keyspace string, hosts ...string) *gocql.ClusterConfig {
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Second,
		Max:        10 * time.Second,
		NumRetries: 5,
	}
	cls := gocql.NewCluster(hosts...)
	cls.Consistency = consistency
	cls.Keyspace = keyspace
	cls.RetryPolicy = retryPolicy
	cls.Timeout = 5 * time.Second
	cls.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	return cls
}
