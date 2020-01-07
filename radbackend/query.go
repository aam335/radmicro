package main

import (
	"context"
	"database/sql"
)

// Prepared contains sql prepared statement, arguments and type
type Prepared struct {
	stmt      *sql.Stmt
	arguments []string
	cacheable bool
}

// CUD = insert/update/delete
func (q *Prepared) CUD(ctx context.Context, attrs map[string]string) (err error) {
	args := []interface{}{}
	for _, attrName := range q.arguments {
		args = append(args, string(attrs[attrName]))
	}
	_, err = q.stmt.ExecContext(ctx, args...)
	return
}

// R = select
func (q *Prepared) R(ctx context.Context, attrs map[string]string) (map[string]string, error) {
	args := []interface{}{}
	for _, attrName := range q.arguments {
		args = append(args, attrs[attrName])
	}

	rows, err := q.stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	outAttrs := make(map[string]string)
	var attrName, attrValue string

	for rows.Next() {
		if err := rows.Scan(&attrName, &attrValue); err != nil {
			rows.Close()
			return nil, err
		}
		outAttrs[attrName] = attrValue
	}

	rerr := rows.Close()
	if rerr != nil {
		return nil, err
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return outAttrs, nil
}

// prepareSQL MAY NOT checks sql syntax
func prepareSQL(c *Config) (db *sql.DB, qs map[string]*Prepared, err error) {
	if db, err = sql.Open(c.SQL.Driver, c.SQL.URI); err != nil {
		return
	}
	qs = make(map[string]*Prepared)
	for topic, q := range c.SQL.Query {
		stmt, err := db.Prepare(q.Prepare)
		if err != nil {
			db.Close()
			return nil, nil, err
		}
		qs[topic] = &Prepared{
			stmt:      stmt,
			cacheable: q.Cacheable,
			arguments: append([]string{}, q.Arguments...),
		}
	}
	return
}
