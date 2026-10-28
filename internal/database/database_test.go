package database_test

import (
	"testing"

	"biblioteca/internal/database"
	"biblioteca/internal/testdb"
)

func TestMigrateIsIdempotent(t *testing.T) {
	url := testdb.URL(t, "test_database")
	if err := database.Migrate(url); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	db, err := database.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for table, want := range map[string]int{"alunos": 5, "livros": 6, "emprestimos": 6} {
		if n := testdb.Count(t, db, "select count(1) from "+table); n != want {
			t.Errorf("%s = %d, want %d", table, n, want)
		}
	}

	var version int
	var dirty bool
	if err := db.QueryRow("select version, dirty from schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if version != 2 || dirty {
		t.Errorf("schema_migrations = %d dirty=%v", version, dirty)
	}
}

func TestOpenUnreachable(t *testing.T) {
	db, err := database.Open("postgres://x:y@127.0.0.1:1/x?sslmode=disable&connect_timeout=2")
	if err == nil {
		db.Close()
		t.Fatal("expected connection error")
	}
}
