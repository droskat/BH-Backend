package connectors

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/droskat/BH-Backend/config"
)

func NewCassandraSession(cfg config.CassandraConfig) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.LocalQuorum
	cluster.Timeout = 5 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	cluster.RetryPolicy = &gocql.ExponentialBackoffRetryPolicy{
		NumRetries: 3,
		Min:        100 * time.Millisecond,
		Max:        2 * time.Second,
	}
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cassandra connection failed: %w", err)
	}
	return session, nil
}
