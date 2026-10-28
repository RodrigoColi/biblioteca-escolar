package cli

import (
	"fmt"
	"strconv"
	"time"

	"biblioteca/internal/emprestimo"
)

func (a *App) listEmprestimos() ([]emprestimo.Emprestimo, error) {
	emprestimos, err := a.Emprestimos.List()
	if err != nil {
		return nil, err
	}
	if len(emprestimos) == 0 {
		fmt.Fprintln(a.out, "Nenhum empréstimo cadastrado.")
		return nil, nil
	}
	now := time.Now()
	rows := make([][]string, len(emprestimos))
	for i, e := range emprestimos {
		rows[i] = []string{strconv.Itoa(e.ID), e.Aluno.Nome, e.Livro.Titulo,
			e.DataEmprestimo.Format(emprestimo.DateLayout), e.Situacao(now)}
	}
	a.table([]string{"ID", "ALUNO", "LIVRO", "EMPRÉSTIMO", "SITUAÇÃO"}, rows)
	return emprestimos, nil
}

func (a *App) insertEmprestimo() error {
	fmt.Fprintln(a.out, "Alunos:")
	if _, err := a.listAlunos(); err != nil {
		return err
	}
	fmt.Fprintln(a.out, "\nLivros:")
	if _, err := a.listLivros(); err != nil {
		return err
	}
	fmt.Fprintln(a.out)

	var e emprestimo.Emprestimo
	e.Aluno.Matricula = a.in.Int("Matrícula do aluno")
	e.Livro.ID = a.in.Int("ID do livro")
	e.DataEmprestimo = a.in.DateOr("Data do empréstimo", today())
	e.DataPrevista = a.in.DateOr("Data prevista de devolução", e.DataEmprestimo.AddDate(0, 0, 14))
	e.DataDevolucao = a.in.OptionalDate("Data de devolução", nil)

	created, err := a.Emprestimos.Create(e)
	if err != nil {
		return err
	}
	a.success("Empréstimo cadastrado:\n  " + created.String())
	return nil
}

func (a *App) updateEmprestimo() error {
	emprestimos, err := a.listEmprestimos()
	if err != nil || len(emprestimos) == 0 {
		return err
	}
	e, err := a.Emprestimos.Get(a.in.Int("\nID do empréstimo a atualizar"))
	if err != nil {
		return err
	}

	fmt.Fprintln(a.out, "\nInforme os novos valores (Enter mantém o valor atual):")
	e.Aluno.Matricula = a.in.IntOr("Matrícula do aluno", e.Aluno.Matricula)
	e.Livro.ID = a.in.IntOr("ID do livro", e.Livro.ID)
	e.DataEmprestimo = a.in.DateOr("Data do empréstimo", e.DataEmprestimo)
	e.DataPrevista = a.in.DateOr("Data prevista de devolução", e.DataPrevista)
	e.DataDevolucao = a.in.OptionalDate("Data de devolução", e.DataDevolucao)

	updated, err := a.Emprestimos.Update(e)
	if err != nil {
		return err
	}
	a.success("Registro atualizado:\n  " + updated.String())
	return nil
}

func (a *App) removeEmprestimo() error {
	emprestimos, err := a.listEmprestimos()
	if err != nil || len(emprestimos) == 0 {
		return err
	}
	e, err := a.Emprestimos.Get(a.in.Int("\nID do empréstimo a remover"))
	if err != nil {
		return err
	}
	fmt.Fprintln(a.out, "\n  "+e.String())
	if !a.in.Confirm("\nConfirma a remoção deste registro?") {
		return errCanceled
	}
	if err := a.Emprestimos.Delete(e.ID); err != nil {
		return err
	}
	a.success("Registro removido com sucesso.")
	return nil
}
