package livro

import (
	"database/sql"
	"errors"
	"testing"

	"biblioteca/internal/testdb"
)

func newPostgres(t *testing.T) (*Postgres, *sql.DB) {
	db := testdb.New(t, "test_livro")
	return NewPostgres(db), db
}

func TestPostgresSeed(t *testing.T) {
	r, _ := newPostgres(t)
	if n, err := r.Count(); err != nil || n != 6 {
		t.Fatalf("Count = %d, %v", n, err)
	}
	all, err := r.FindAll()
	if err != nil || len(all) != 6 {
		t.Fatalf("FindAll = %d, %v", len(all), err)
	}
	for i, l := range all {
		if l.ID != i+1 {
			t.Errorf("FindAll should be ordered by id: %v", all)
			break
		}
	}
}

func TestPostgresCRUD(t *testing.T) {
	r, _ := newPostgres(t)
	l := Livro{Titulo: "O'Reilly's Go", Autor: "Fulano", Genero: "Técnico", AnoPublicacao: 2020}

	id, err := r.Insert(l)
	if err != nil {
		t.Fatal(err)
	}
	if id != 7 {
		t.Errorf("identity id = %d, want 7", id)
	}
	l.ID = id
	if got, err := r.FindByID(id); err != nil || got != l {
		t.Fatalf("FindByID = %+v, %v", got, err)
	}

	l.Titulo, l.AnoPublicacao = "Go Atualizado", 2021
	if err := r.Update(l); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.FindByID(id); got != l {
		t.Errorf("after update = %+v", got)
	}

	if err := r.Delete(id); err != nil {
		t.Fatal(err)
	}
	if _, err := r.FindByID(id); !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID after delete = %v, want ErrNotFound", err)
	}
	if ok, _ := r.Exists(id); ok {
		t.Error("deleted livro should not exist")
	}
}

func TestPostgresCountEmprestimos(t *testing.T) {
	r, _ := newPostgres(t)
	for id, want := range map[int]int{1: 1, 4: 1, 99: 0} {
		if n, err := r.CountEmprestimos(id); err != nil || n != want {
			t.Errorf("CountEmprestimos(%d) = %d, %v; want %d", id, n, err, want)
		}
	}
}

func TestPostgresDeleteWithLoansViolatesFK(t *testing.T) {
	r, _ := newPostgres(t)
	if err := r.Delete(1); err == nil {
		t.Error("expected foreign key violation")
	}
	if ok, _ := r.Exists(1); !ok {
		t.Error("livro should still exist")
	}
}

func TestPostgresDeleteWithEmprestimos(t *testing.T) {
	r, db := newPostgres(t)
	if err := r.DeleteWithEmprestimos(1); err != nil {
		t.Fatal(err)
	}
	if n := testdb.Count(t, db, "select count(1) from livros"); n != 5 {
		t.Errorf("livros = %d, want 5", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 5 {
		t.Errorf("emprestimos = %d, want 5", n)
	}
}
