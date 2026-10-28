// Package testdb prepares an isolated, migrated PostgreSQL schema for integration tests.
package testdb

import (
	"database/sql"
	"net/url"
	"os"
	"testing"

	"biblioteca/internal/database"
)

const EnvVar = "TEST_DATABASE_URL"

// URL recreates the given schema, applies the migrations (including the seed) and
// returns a connection URL scoped to it. The test is skipped when EnvVar is unset.
func URL(t testing.TB, schema string) string {
	t.Helper()
	base := os.Getenv(EnvVar)
	if base == "" {
		t.Skip(EnvVar + " não definida; teste de integração ignorado")
	}

	admin, err := database.Open(base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if _, err := admin.Exec("drop schema if exists " + schema + " cascade"); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("create schema " + schema); err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()

	if err := database.Migrate(u.String()); err != nil {
		t.Fatal(err)
	}
	return u.String()
}

func New(t testing.TB, schema string) *sql.DB {
	t.Helper()
	db, err := database.Open(URL(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func Exec(t testing.TB, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
}

func Count(t testing.TB, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}
