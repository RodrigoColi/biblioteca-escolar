package emprestimo

import (
	"errors"
	"fmt"
	"time"

	"biblioteca/internal/aluno"
	"biblioteca/internal/livro"
)

const DateLayout = "02/01/2006"

var ErrNotFound = errors.New("empréstimo não encontrado")

type Emprestimo struct {
	ID             int
	Aluno          aluno.Aluno
	Livro          livro.Livro
	DataEmprestimo time.Time
	DataPrevista   time.Time
	DataDevolucao  *time.Time
}

func (e Emprestimo) Situacao(agora time.Time) string {
	y, m, d := agora.Date()
	hoje := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	switch {
	case e.DataDevolucao != nil:
		return "Devolvido"
	case e.DataPrevista.Before(hoje):
		return "Atrasado"
	default:
		return "Em aberto"
	}
}

func (e Emprestimo) String() string {
	devolucao := "-"
	if e.DataDevolucao != nil {
		devolucao = e.DataDevolucao.Format(DateLayout)
	}
	return fmt.Sprintf("ID: %d | Aluno: %s (%d) | Livro: %s (%d) | Empréstimo: %s | Previsão: %s | Devolução: %s | Situação: %s",
		e.ID, e.Aluno.Nome, e.Aluno.Matricula, e.Livro.Titulo, e.Livro.ID,
		e.DataEmprestimo.Format(DateLayout), e.DataPrevista.Format(DateLayout), devolucao, e.Situacao(time.Now()))
}
