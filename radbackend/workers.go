package main

import (
	"context"
	"database/sql"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
	"golang.org/x/time/rate"
)

type workerType = func(db *sql.DB, nc *nats.Conn, rc *Cache, workerID int, donechan chan struct{}) error

func worker(db *sql.DB, nc *nats.Conn, workerID int) error {

	return nil
}

func runWorkers(ctx context.Context, c *Config, rc *Cache, w workerType) {
	l := rate.NewLimiter(rate.Every(c.Server.RestartInterval.Duration), c.Server.MaxConnections)
	totalConn := 0
	exitChan, doneChan := make(chan struct{}), make(chan struct{})
	workerid := 0
	wg := sync.WaitGroup{}
	defer func() {
		close(doneChan)
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
			var nc *nats.Conn
			var db *sql.DB
			if db, err = sql.Open(c.SQL.Driver, c.SQL.URI); err != nil {
				return
			}
			defer db.Close()
			if nc, err = nats.Connect(c.Nats.URI); err != nil {
				return
			}
			defer nc.Close()

			if err = w(db, nc, rc, workerid, doneChan); err != nil {
				log.Printf("Error [#%v] %v", workerid, err)
			}
		}(workerid)
		workerid++
	}
}
