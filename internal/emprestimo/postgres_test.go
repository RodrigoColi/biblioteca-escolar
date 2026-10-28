package emprestimo

import (
	"database/sql"
	"errors"
	"testing"

	"biblioteca/internal/testdb"
)

func newPostgres(t *testing.T) (*Postgres, *sql.DB) {
	db := testdb.New(t, "test_emprestimo")
	return NewPostgres(db), db
}

func TestPostgresFindLoadsAssociations(t *testing.T) {
	r, _ := newPostgres(t)
	all, err := r.FindAll()
	if err != nil || len(all) != 6 {
		t.Fatalf("FindAll = %d, %v", len(all), err)
	}

	e, err := r.FindByID(1)
	if err != nil {
		t.Fatal(err)
	}
	if e.Aluno.Nome != "Ana Souza" || e.Aluno.Email != "ana.souza@escola.com" || e.Livro.Titulo != "Dom Casmurro" || e.Livro.AnoPublicacao != 1899 {
		t.Errorf("associations not loaded: %+v", e)
	}
	if !e.DataEmprestimo.Equal(day(2026, 9, 1)) || !e.DataPrevista.Equal(day(2026, 9, 15)) {
		t.Errorf("dates = %v, %v", e.DataEmprestimo, e.DataPrevista)
	}
	if e.DataDevolucao == nil || !e.DataDevolucao.Equal(day(2026, 9, 14)) {
		t.Errorf("DataDevolucao = %v", e.DataDevolucao)
	}

	if open, _ := r.FindByID(3); open.DataDevolucao != nil {
		t.Errorf("loan 3 should be open: %v", open.DataDevolucao)
	}
	if _, err := r.FindByID(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID(99) = %v, want ErrNotFound", err)
	}
}

func TestPostgresCRUD(t *testing.T) {
	r, _ := newPostgres(t)
	e := Emprestimo{DataEmprestimo: day(2026, 10, 1), DataPrevista: day(2026, 10, 15)}
	e.Aluno.Matricula = 1002
	e.Livro.ID = 1

	id, err := r.Insert(e)
	if err != nil {
		t.Fatal(err)
	}
	if id != 7 {
		t.Errorf("identity id = %d, want 7", id)
	}
	got, err := r.FindByID(id)
	if err != nil || got.Aluno.Nome != "Bruno Lima" || got.DataDevolucao != nil {
		t.Fatalf("FindByID = %+v, %v", got, err)
	}

	devolucao := day(2026, 10, 10)
	got.DataDevolucao = &devolucao
	got.Livro.ID = 3
	if err := r.Update(got); err != nil {
		t.Fatal(err)
	}
	updated, _ := r.FindByID(id)
	if updated.DataDevolucao == nil || !updated.DataDevolucao.Equal(devolucao) || updated.Livro.Titulo != "O Pequeno Príncipe" {
		t.Errorf("after update = %+v", updated)
	}

	updated.DataDevolucao = nil
	if err := r.Update(updated); err != nil {
		t.Fatal(err)
	}
	if cleared, _ := r.FindByID(id); cleared.DataDevolucao != nil {
		t.Errorf("DataDevolucao should be cleared: %v", cleared.DataDevolucao)
	}

	if err := r.Delete(id); err != nil {
		t.Fatal(err)
	}
	if ok, _ := r.Exists(id); ok {
		t.Error("deleted loan should not exist")
	}
}

func TestPostgresCountOpenByLivro(t *testing.T) {
	r, _ := newPostgres(t)
	tests := []struct {
		livro, ignore, want int
	}{
		{4, 0, 1},
		{4, 3, 0},
		{1, 0, 0},
		{99, 0, 0},
	}
	for _, tt := range tests {
		if n, err := r.CountOpenByLivro(tt.livro, tt.ignore); err != nil || n != tt.want {
			t.Errorf("CountOpenByLivro(%d, %d) = %d, %v; want %d", tt.livro, tt.ignore, n, err, tt.want)
		}
	}
}

func TestPostgresConstraints(t *testing.T) {
	r, db := newPostgres(t)

	invalidDates := Emprestimo{DataEmprestimo: day(2026, 10, 15), DataPrevista: day(2026, 10, 1)}
	invalidDates.Aluno.Matricula, invalidDates.Livro.ID = 1001, 1
	if _, err := r.Insert(invalidDates); err == nil {
		t.Error("expected check constraint violation for data_prevista")
	}

	before := day(2026, 9, 1)
	invalidReturn := Emprestimo{DataEmprestimo: day(2026, 10, 1), DataPrevista: day(2026, 10, 15), DataDevolucao: &before}
	invalidReturn.Aluno.Matricula, invalidReturn.Livro.ID = 1001, 1
	if _, err := r.Insert(invalidReturn); err == nil {
		t.Error("expected check constraint violation for data_devolucao")
	}

	missingAluno := Emprestimo{DataEmprestimo: day(2026, 10, 1), DataPrevista: day(2026, 10, 15)}
	missingAluno.Aluno.Matricula, missingAluno.Livro.ID = 9999, 1
	if _, err := r.Insert(missingAluno); err == nil {
		t.Error("expected foreign key violation")
	}

	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 6 {
		t.Errorf("emprestimos = %d, want 6", n)
	}
}
