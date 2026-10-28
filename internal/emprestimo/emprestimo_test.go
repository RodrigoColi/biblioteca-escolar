package emprestimo

import (
	"strings"
	"testing"
	"time"

	"biblioteca/internal/aluno"
	"biblioteca/internal/livro"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestSituacao(t *testing.T) {
	agora := time.Date(2026, 10, 28, 15, 30, 0, 0, time.Local)
	devolvido := day(2026, 10, 20)

	tests := []struct {
		name string
		e    Emprestimo
		want string
	}{
		{"devolvido mesmo após o prazo", Emprestimo{DataPrevista: day(2026, 10, 1), DataDevolucao: &devolvido}, "Devolvido"},
		{"atrasado", Emprestimo{DataPrevista: day(2026, 10, 27)}, "Atrasado"},
		{"vence hoje", Emprestimo{DataPrevista: day(2026, 10, 28)}, "Em aberto"},
		{"em aberto", Emprestimo{DataPrevista: day(2026, 11, 10)}, "Em aberto"},
	}
	for _, tt := range tests {
		if got := tt.e.Situacao(agora); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestString(t *testing.T) {
	e := Emprestimo{
		ID:             3,
		Aluno:          aluno.Aluno{Matricula: 1001, Nome: "Ana"},
		Livro:          livro.Livro{ID: 7, Titulo: "Dom Casmurro"},
		DataEmprestimo: day(2026, 9, 1),
		DataPrevista:   day(2026, 9, 15),
	}
	s := e.String()
	for _, want := range []string{"ID: 3", "Ana (1001)", "Dom Casmurro (7)", "01/09/2026", "15/09/2026", "Devolução: -", "Atrasado"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, missing %q", s, want)
		}
	}

	d := day(2026, 9, 14)
	e.DataDevolucao = &d
	if s := e.String(); !strings.Contains(s, "Devolução: 14/09/2026") || !strings.Contains(s, "Devolvido") {
		t.Errorf("String() = %q", s)
	}
}
