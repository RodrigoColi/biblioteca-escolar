package aluno

import (
	"database/sql"
	"errors"
	"testing"

	"biblioteca/internal/testdb"
)

func newPostgres(t *testing.T) (*Postgres, *sql.DB) {
	db := testdb.New(t, "test_aluno")
	return NewPostgres(db), db
}

func TestPostgresSeed(t *testing.T) {
	r, _ := newPostgres(t)
	n, err := r.Count()
	if err != nil || n != 5 {
		t.Fatalf("Count = %d, %v", n, err)
	}
	all, err := r.FindAll()
	if err != nil || len(all) != 5 {
		t.Fatalf("FindAll = %d, %v", len(all), err)
	}
	if all[0].Nome != "Ana Souza" || all[4].Nome != "Eduarda Rocha" {
		t.Errorf("FindAll should be ordered by nome: %v", all)
	}
}

func TestPostgresCRUD(t *testing.T) {
	r, _ := newPostgres(t)
	a := Aluno{Matricula: 2000, Nome: "Fulano O'Neil", Email: "fulano@escola.com", Turma: "6A"}

	if err := r.Insert(a); err != nil {
		t.Fatal(err)
	}
	if ok, _ := r.Exists(2000); !ok {
		t.Fatal("inserted aluno should exist")
	}
	got, err := r.FindByID(2000)
	if err != nil || got != a {
		t.Fatalf("FindByID = %+v, %v", got, err)
	}

	a.Nome, a.Email, a.Turma = "Fulano Atualizado", "novo@escola.com", "7B"
	if err := r.Update(a); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.FindByID(2000); got != a {
		t.Errorf("after update = %+v", got)
	}

	if err := r.Delete(2000); err != nil {
		t.Fatal(err)
	}
	if ok, _ := r.Exists(2000); ok {
		t.Error("deleted aluno should not exist")
	}
	if _, err := r.FindByID(2000); !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID after delete = %v, want ErrNotFound", err)
	}
}

func TestPostgresInsertDuplicateFails(t *testing.T) {
	r, _ := newPostgres(t)
	if err := r.Insert(Aluno{Matricula: 1001, Nome: "X", Email: "x@x.com", Turma: "1A"}); err == nil {
		t.Error("expected primary key violation")
	}
}

func TestPostgresStoresInjectionAttemptLiterally(t *testing.T) {
	r, db := newPostgres(t)
	nome := "Robert'); DROP TABLE alunos;--"
	if err := r.Insert(Aluno{Matricula: 2001, Nome: nome, Email: "r@x.com", Turma: "1A"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.FindByID(2001); got.Nome != nome {
		t.Errorf("nome = %q", got.Nome)
	}
	if n := testdb.Count(t, db, "select count(1) from alunos"); n != 6 {
		t.Errorf("alunos = %d, want 6", n)
	}
}

func TestPostgresCountEmprestimos(t *testing.T) {
	r, _ := newPostgres(t)
	for matricula, want := range map[int]int{1001: 2, 1004: 1, 9999: 0} {
		if n, err := r.CountEmprestimos(matricula); err != nil || n != want {
			t.Errorf("CountEmprestimos(%d) = %d, %v; want %d", matricula, n, err, want)
		}
	}
}

func TestPostgresDeleteWithLoansViolatesFK(t *testing.T) {
	r, _ := newPostgres(t)
	if err := r.Delete(1001); err == nil {
		t.Error("expected foreign key violation")
	}
	if ok, _ := r.Exists(1001); !ok {
		t.Error("aluno should still exist")
	}
}

func TestPostgresDeleteWithEmprestimos(t *testing.T) {
	r, db := newPostgres(t)
	if err := r.DeleteWithEmprestimos(1001); err != nil {
		t.Fatal(err)
	}
	if n := testdb.Count(t, db, "select count(1) from alunos"); n != 4 {
		t.Errorf("alunos = %d, want 4", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 4 {
		t.Errorf("emprestimos = %d, want 4", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos where matricula = 1001"); n != 0 {
		t.Errorf("emprestimos of 1001 = %d, want 0", n)
	}
}
