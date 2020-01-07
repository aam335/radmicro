package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
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
	w := func(ctx context.Context, c *Config, rc *Cache, workerID int) error {
		m.Lock()
		count++
		m.Unlock()
		<-ctx.Done()
		return nil
	}
	ew := func(ctx context.Context, c *Config, rc *Cache, workerID int) error {
		m.Lock()
		count++
		m.Unlock()
		return errors.New("force error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	c.runWorkers(ctx, rc, w)
	cancel()
	if count != 10 {
		t.Error("Burst error")
	}
	count = 0
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	c.runWorkers(ctx, rc, ew)
	cancel()
	if count <= 10 || count >= (10+1000/50) {
		t.Error("Rate error")
	}

}

func TestConfig_prepareSQL(t *testing.T) {
	sqlURI := "file::memory:?cache=shared"
	sqlDriver := "sqlite3"
	db := newDb("?cache=shared")
	defer db.Close()

	tests := []struct {
		name    string
		Query   map[string]Query
		topic   string
		db      bool
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "nil db", Query: map[string]Query{"Auth": Query{Prepare: "select * from users", Cacheable: true}}, wantErr: true},
		{name: "select w/o args", topic: "Auth", db: true, Query: map[string]Query{"Auth": Query{Prepare: "select count(*) from users", Cacheable: true}}, wantErr: false},
		{name: "select w/args", topic: "Auth", db: true, Query: map[string]Query{"Auth": Query{Prepare: "select * from users where user=:1", Cacheable: false}}, wantErr: true},
		{name: "select w/args, errored syntax", topic: "Auth", db: true, Query: map[string]Query{"Auth": Query{Prepare: "select * from users where user=:1 jhgjhgjh", Cacheable: false}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{}
			if tt.db {
				c.SQL.Driver = sqlDriver
				c.SQL.URI = sqlURI
			}
			c.SQL.Query = tt.Query
			_, _, err := prepareSQL(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.prepareSQL() error = %v, wantErr %v", err, tt.wantErr)
				//				return
			}
		})
	}
}
