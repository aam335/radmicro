package main

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
	"golang.org/x/time/rate"
)

type workerType = func(c *Config, rc *Cache, workerID int, donechan <-chan struct{}) error

func worker(c *Config, rc *Cache, workerID int, doneChan <-chan struct{}) (err error) {
	var nc *nats.Conn
	var db *sql.DB
	var qs map[string]*query

	if nc, err = nats.Connect(c.Nats.URI); err != nil {
		return
	}
	defer nc.Close()

	if db, qs, err = c.prepareSQL(); err != nil {
		return
	}
	defer db.Close()
	subject := c.Server.ServiceName + ".req.*" //
	for {

	}

	return nil
}

type query struct {
	stmt      *sql.Stmt
	arguments []string
	cacheable bool
}

// prepareSQL MAY NOT checks sql syntax
func (c *Config) prepareSQL() (db *sql.DB, qs map[string]*query, err error) {
	if db, err = sql.Open(c.SQL.Driver, c.SQL.URI); err != nil {
		return
	}
	qs = make(map[string]*query)
	for topic, q := range c.SQL.Query {
		stmt, err := db.Prepare(q.Prepare)
		if err != nil {
			db.Close()
			return nil, nil, err
		}
		qs[topic] = &query{
			stmt:      stmt,
			cacheable: q.Cacheable,
			arguments: append([]string{}, q.Arguments...),
		}
	}
	return
}

// ExecArgs converts map into []string slice as described in Query.Arguments
func (q *Query) ExecArgs(args map[string]string) (argsSlice []string) {
	for _, attrName := range q.Arguments {
		argsSlice = append(argsSlice, args[attrName])
	}
	return
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

			if err = w(c, rc, workerid, ctx.Done()); err != nil {
				log.Printf("Error [#%v] %v", workerid, err)
			}
		}(workerid)
		workerid++
	}
}
