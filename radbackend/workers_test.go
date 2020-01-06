package main

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func Test_runWorkers(t *testing.T) {
	s, natsURI := runNatsInstance()
	defer s.Shutdown()
	miniRedis, err := miniredis.Run()
	require.NoError(t, err)
	defer miniRedis.Close()
	// db := newDb("?cache=shared")
	// defer db.Close()

	c := Config{}
	c.Nats.URI = natsURI
	c.SQL.Driver = "sqlite3"
	c.SQL.URI = "file::memory:?cache=shared"
	c.Server.MaxConnections = 10
	c.Server.RestartInterval.Duration = time.Millisecond * 50
	c.Redis.URI = miniRedis.Addr()
	rc, err := NewCache(c.Redis.URI, c.Server.ServiceName, 5)
	require.NoError(t, err)

	count := 0
	m := sync.Mutex{}
	w := func(db *sql.DB, nc *nats.Conn, rc *Cache, workerID int, c chan struct{}) error {
		m.Lock()
		count++
		m.Unlock()
		<-c
		return nil
	}
	ew := func(db *sql.DB, nc *nats.Conn, rc *Cache, workerID int, c chan struct{}) error {
		m.Lock()
		count++
		m.Unlock()
		return errors.New("force error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	runWorkers(ctx, &c, rc, w)
	cancel()
	if count != 10 {
		t.Error("Burst error")
	}
	count = 0
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	runWorkers(ctx, &c, rc, ew)
	cancel()
	if count <= 10 || count >= (10+1000/50) {
		t.Error("Rate error")
	}

}
