package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, int(10), count, "Burst error")
	count = 0
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	c.runWorkers(ctx, rc, ew)
	cancel()
	assert.False(t, count <= 10 || count >= (10+1000/50), "Rate error count=", count)
}

func TestPrepareSQL(t *testing.T) {
	sqlDriver := "sqlite3"
	db, sqlURI := newDb("prepareSQL")
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
			}
		})
	}
}

func TestWorker(t *testing.T) {
	s, natsURI := runNatsInstance()
	defer s.Shutdown()
	miniRedis, err := miniredis.Run()
	require.NoError(t, err)
	defer miniRedis.Close()
	db, sqlURI := newDb("worker")
	defer db.Close()
	// natsURI = nats.DefaultURL
	c := Config{}
	c.Nats.URI = natsURI
	c.Nats.QueueName = "queue"
	c.SQL.Driver = "sqlite3"
	c.SQL.URI = sqlURI
	c.Server.ServiceName = "inet"
	c.Redis.URI = miniRedis.Addr()
	c.Redis.TTLDefault.Duration = time.Minute
	c.SQL.Query = map[string]Query{
		"Auth": {Prepare: "select attrname,attrvalue from users where username=:1",
			Arguments: []string{"user"}, Cacheable: true},
		"Start": {Prepare: "insert into acc(username, sessionid) values(:1,:2)",
			Arguments: []string{"user", "session"}},
	}

	rc, err := NewCache(c.Redis.URI, c.Server.ServiceName, 5)
	require.NoError(t, err)
	errCh := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		err := worker(ctx, &c, rc, 555)
		if err != context.DeadlineExceeded {
			t.Error(err)
		}
		errCh <- err
	}()
	defer cancel()
	wg.Wait()
	time.Sleep(500 * time.Millisecond) // waits worker start
	nc, err := nats.Connect(c.Nats.URI)
	require.NoError(t, err)
	// Auth
	topic := c.Server.ServiceName + ".req.Auth"
	msg, err := nc.RequestWithContext(ctx, topic, []byte(`{"user":"ut"}`))
	require.NoError(t, err)
	ret := make(map[string]string)
	require.NoError(t, json.Unmarshal(msg.Data, &ret))
	require.Equal(t, map[string]string{"attr1": "val1", "attr2": "val2"}, ret)
	// Start
	topic = c.Server.ServiceName + ".req.Start"
	err = nc.Publish(topic, []byte(`{"user":"ut","session":"sessionID1234567890"}`))
	// todo:request from cache!!!

	//
	select {
	case err := <-errCh:
		if err != context.DeadlineExceeded {
			require.NoError(t, err)
		}
	case <-time.After(time.Second * 2):
		t.Error("exit worker on ctx.Done() fail")
	}
	res := dumpSQL(db, "select username,sessionid from acc where username='ut'")
	require.Equal(t, res, []string{"username;sessionid", "ut;sessionID1234567890;"})
}
