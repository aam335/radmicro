package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	server "github.com/nats-io/nats-server/v2/server"
)

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
func newDb(file string) (db *sql.DB, uri string) {
	uri = "file:" + file + "?mode=memory&cache=shared"
	db, err := sql.Open("sqlite3", uri)
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

func dumpSQL(db *sql.DB, sqlQuery string) []string {
	rows, err := db.Query(sqlQuery)
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
	ret := []string{strings.Join(cols, ";")}
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
		// log.Println(val)
		ret = append(ret, val)
	}
	return ret
}
