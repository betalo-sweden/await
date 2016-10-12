package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"context"

	_ "github.com/lib/pq" // Register Postgres driver
)

type postgresqlResource struct {
	url.URL
}

func (r *postgresqlResource) Await(ctx context.Context) error {
	tags := parseTags(r.URL.Fragment)

	dbParams := url.Values{}

	if val, ok := tags["ssl"]; ok && val == "false" {
		dbParams.Set("sslmode", "disable") // PGSSLMODE
	}

	dsnURL := r.URL
	dsnURL.Fragment = ""
	dsnURL.RawQuery = dbParams.Encode()
	dsn := dsnURL.String()

	db, err := sql.Open(r.URL.Scheme, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return &unavailableError{err}
	}

	if val, ok := tags["tables"]; ok {
		tables := strings.Split(val, ",")
		if err := awaitPostgreSQLTables(db, dsnURL.Path[1:], tables); err != nil {
		}
	}

	return nil
}

func awaitPostgreSQLTables(db *sql.DB, dbName string, tables []string) error {
	if len(tables) == 0 {
		const stmt = `SELECT count(*) FROM information_schema.tables WHERE table_catalog=? AND table_schema='public'`
		var tableCnt int
		if err := db.QueryRow(stmt, dbName).Scan(&tableCnt); err != nil {
			return err
		}

		if tableCnt == 0 {
			return &unavailableError{errors.New("no tables found")}
		}

		return nil
	}

	const stmt = `SELECT table_name FROM information_schema.tables WHERE table_catalog=? AND table_schema='public'`
	rows, err := db.Query(stmt, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

	var actualTables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return err
		}
		actualTables = append(actualTables, t)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	contains := func(l []string, s string) bool {
		for _, i := range l {
			if i == s {
				return true
			}
		}
		return false
	}

	for _, t := range tables {
		if !contains(actualTables, t) {
			return &unavailableError{fmt.Errorf("table not found: %s", t)}
		}
	}

	return nil
}
