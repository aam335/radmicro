package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats-server/server"
)

type sqlUser struct {
	name      string
	attrname  string
	attrvalue string
}

var users = []sqlUser{
	{name: "00:00:00:00:00:00", attrname: "Acct-Interim-Interval", attrvalue: "600"},
	{name: "00:00:00:00:00:00", attrname: "Framed-IP-Address", attrvalue: "192.168.2.100"},
	{name: "00:00:00:00:00:00", attrname: "Framed-IP-Netmask", attrvalue: "255.255.255.0"},
	{name: "00:00:00:00:00:00", attrname: "Session-Timeout", attrvalue: "600"},
}

func runNatsInstance() (*server.Server, string) {
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           server.RANDOM_PORT,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 2048,
	}
	s, err := server.NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(5 * time.Second) {
		log.Fatal("Unable to start NATS Server in Go Routine")
	}
	uri := s.Addr().String()
	return s, uri
}

// ?cache=shared
func newDb(shared string) (db *sql.DB) {
	db, err := sql.Open("sqlite3", "file::memory:"+shared)
	if err != nil {
		log.Fatal(err)
	}
	sqlStmt, err := ioutil.ReadFile("structs.sql")
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
	_, err = db.Exec(string(sqlStmt))
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
	prep, err := db.Prepare("insert into users(name,attrname,attrvalue) values (:1,:2,:3)")
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range users {
		if _, err = prep.Exec(q.name, q.attrname, q.attrvalue); err != nil {
			log.Fatal(err)
		}
	}

	return
}
