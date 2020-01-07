package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats-server/server"
)

type sqlUser struct {
	name      string
	attrname  string
	attrvalue string
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

	return
}

func dumpDB(db *sql.DB, tablename string) {
	rows, err := db.Query("select * from " + tablename)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	var strs = make([]sql.NullString, len(cols))
	var ptrs = make([]interface{}, len(cols))
	for i := range strs {
		ptrs[i] = &strs[i]
	}
	log.Println(strings.Join(cols, ";"))
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			log.Fatal(err)
		}
		val := ""
		for _, v := range strs {
			if v.Valid {
				val += v.String
			} else {
				val += "NULL"
			}
			val += ";"
		}
		log.Println(val)
	}

}
