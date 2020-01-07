package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"golang.org/x/time/rate"
)

type workerType = func(ctx context.Context, c *Config, rc *Cache, workerID int) error

func worker(ctx context.Context, c *Config, rc *Cache, workerID int) (err error) {
	var nc *nats.Conn
	var db *sql.DB
	var qs map[string]*Prepared

	if nc, err = nats.Connect(c.Nats.URI); err != nil {
		return
	}
	defer nc.Close()

	if db, qs, err = prepareSQL(c); err != nil {
		return
	}
	defer db.Close()

	subject := c.Server.ServiceName + ".req.*" //
	ch := make(chan *nats.Msg, 64)
	sub, err := nc.ChanSubscribe(subject, ch)
	defer sub.Unsubscribe()
	var msg *nats.Msg

	for {
		select {
		case msg = <-ch:
		case <-ctx.Done():
			return
		}

		topic := msg.Subject[len(subject)-1:]
		query := qs[topic]
		// Not known topic
		if query == nil {
			continue
		}
		attrs := make(map[string]string)
		if err == json.Unmarshal(msg.Data, &attrs) {
			log.Printf("Wrong encoded message on %v `%v`", msg.Subject, string(msg.Data))
			continue
		}

		if query.cacheable == false {
			if err = query.CUD(ctx, attrs); err != nil {
				return
			}
			continue
		}
		// cacheable Auth
		var data []byte
		data, err = rc.GetCache(msg.Reply, func(key string) (ttl int, data []byte, err error) {
			var rattrs map[string]string
			if rattrs, err = query.R(ctx, attrs); err != nil {
				return
			}

			if strTTL, ok := rattrs[c.Redis.TTLAttr]; ok {
				dTTL, _ := time.ParseDuration(strTTL)
				ttl = int(dTTL.Seconds())
			} else {
				ttl = int(c.Redis.TTLDefault.Seconds())
			}

			data, err = json.Marshal(rattrs)
			return
		})
		if err != nil {
			return
		}
		if err = msg.Respond(data); err != nil {
			return
		}
	}
}

func (c *Config) runWorkers(ctx context.Context, rc *Cache, w workerType) {
	l := rate.NewLimiter(rate.Every(c.Server.RestartInterval.Duration), c.Server.MaxConnections)
	totalConn := 0
	exitChan := make(chan struct{})
	workerid := 0
	wg := sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	for {
		if err := l.Wait(ctx); err != nil {
			return
		}
		if totalConn >= c.Server.MaxConnections {
			select {
			case <-exitChan:
			case <-ctx.Done():
				return
			}
			totalConn--
		}
		totalConn++
		log.Printf("run worker [#%v] %v/%v", workerid, totalConn, c.Server.MaxConnections)
		wg.Add(1)
		go func(workerid int) {
			var err error
			defer func() {
				wg.Done()
				log.Printf("Error [#%v] %v", workerid, err)
				exitChan <- struct{}{}
			}()

			if err = w(ctx, c, rc, workerid); err != nil {
				log.Printf("Error [#%v] %v", workerid, err)
			}
		}(workerid)
		workerid++
	}
}
