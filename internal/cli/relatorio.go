package cli

import (
	"fmt"
	"strconv"

	"biblioteca/internal/emprestimo"
)

func (a *App) reportsMenu() {
	for {
		a.header("RELATÓRIOS")
		fmt.Fprintln(a.out, " 1 - Empréstimos por gênero")
		fmt.Fprintln(a.out, " 2 - Empréstimos detalhados (aluno x livro)")
		fmt.Fprintln(a.out, " 0 - Voltar ao menu principal")

		var err error
		switch a.in.Option("\nEscolha um relatório", 0, 2) {
		case 0:
			return
		case 1:
			a.header("RELATÓRIOS » EMPRÉSTIMOS POR GÊNERO")
			err = a.reportPorGenero()
		case 2:
			a.header("RELATÓRIOS » EMPRÉSTIMOS DETALHADOS")
			err = a.reportDetalhado()
		}
		if err != nil {
			a.fail(err)
		}
		a.in.Pause()
	}
}

func (a *App) reportPorGenero() error {
	data, err := a.Relatorios.EmprestimosPorGenero()
	if err != nil {
		return err
	}
	rows := make([][]string, len(data))
	for i, g := range data {
		rows[i] = []string{g.Genero, strconv.Itoa(g.Total), strconv.Itoa(g.EmAberto), strconv.Itoa(g.Atrasados)}
	}
	a.table([]string{"GÊNERO", "TOTAL", "EM ABERTO", "ATRASADOS"}, rows)
	return nil
}

func (a *App) reportDetalhado() error {
	data, err := a.Relatorios.EmprestimosDetalhados()
	if err != nil {
		return err
	}
	if len(data) == 0 {
		fmt.Fprintln(a.out, "Nenhum empréstimo cadastrado.")
		return nil
	}
	rows := make([][]string, len(data))
	for i, d := range data {
		devolucao := "-"
		if d.DataDevolucao != nil {
			devolucao = d.DataDevolucao.Format(emprestimo.DateLayout)
		}
		rows[i] = []string{strconv.Itoa(d.ID), d.Aluno, d.Turma, d.Livro, d.Autor,
			d.DataEmprestimo.Format(emprestimo.DateLayout), d.DataPrevista.Format(emprestimo.DateLayout), devolucao, d.Situacao}
	}
	a.table([]string{"ID", "ALUNO", "TURMA", "LIVRO", "AUTOR", "EMPRÉSTIMO", "PREVISÃO", "DEVOLUÇÃO", "SITUAÇÃO"}, rows)
	return nil
}
