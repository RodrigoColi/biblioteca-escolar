package relatorio

import (
	"testing"

	"biblioteca/internal/testdb"
)

// Loans are relative to current_date so the expected situação never depends on when the test runs.
func newPostgres(t *testing.T) *Postgres {
	db := testdb.New(t, "test_relatorio")
	testdb.Exec(t, db, "truncate emprestimos")
	testdb.Exec(t, db, `insert into emprestimos (matricula, id_livro, data_emprestimo, data_prevista, data_devolucao) values
		(1001, 1, current_date - 30, current_date - 16, current_date - 20),
		(1002, 2, current_date - 30, current_date - 10, null),
		(1003, 4, current_date - 2,  current_date + 12, null),
		(1004, 3, current_date - 14, current_date,      null)`)
	return NewPostgres(db)
}

func TestEmprestimosPorGenero(t *testing.T) {
	got, err := newPostgres(t).EmprestimosPorGenero()
	if err != nil {
		t.Fatal(err)
	}
	want := []PorGenero{
		{Genero: "Romance", Total: 2, EmAberto: 1, Atrasados: 1},
		{Genero: "Ficção", Total: 1, EmAberto: 1, Atrasados: 0},
		{Genero: "Infantojuvenil", Total: 1, EmAberto: 1, Atrasados: 0},
		{Genero: "Aventura", Total: 0, EmAberto: 0, Atrasados: 0},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestEmprestimosDetalhados(t *testing.T) {
	got, err := newPostgres(t).EmprestimosDetalhados()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct{ aluno, situacao string }{
		"A Revolução dos Bichos": {"Carla Mendes", "Em aberto"},
		"O Pequeno Príncipe":     {"Diego Ferreira", "Em aberto"},
		"Dom Casmurro":           {"Ana Souza", "Devolvido"},
		"O Cortiço":              {"Bruno Lima", "Atrasado"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows", len(got))
	}
	if got[0].Livro != "A Revolução dos Bichos" {
		t.Errorf("rows should be ordered by data_emprestimo desc, first = %q", got[0].Livro)
	}
	for _, d := range got {
		w, ok := want[d.Livro]
		if !ok || d.Aluno != w.aluno || d.Situacao != w.situacao {
			t.Errorf("row %+v, want %+v", d, w)
		}
		if (d.Situacao == "Devolvido") != (d.DataDevolucao != nil) {
			t.Errorf("DataDevolucao inconsistent with situação: %+v", d)
		}
	}
}

func TestRelatoriosSemEmprestimos(t *testing.T) {
	db := testdb.New(t, "test_relatorio")
	testdb.Exec(t, db, "truncate emprestimos")
	r := NewPostgres(db)

	detalhado, err := r.EmprestimosDetalhados()
	if err != nil || len(detalhado) != 0 {
		t.Errorf("EmprestimosDetalhados = %+v, %v", detalhado, err)
	}
	porGenero, err := r.EmprestimosPorGenero()
	if err != nil || len(porGenero) != 4 {
		t.Fatalf("EmprestimosPorGenero = %+v, %v", porGenero, err)
	}
	for _, g := range porGenero {
		if g.Total != 0 || g.EmAberto != 0 || g.Atrasados != 0 {
			t.Errorf("expected zeros: %+v", g)
		}
	}
}
