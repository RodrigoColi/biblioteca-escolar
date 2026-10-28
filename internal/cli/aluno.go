package cli

import (
	"fmt"
	"strconv"

	"biblioteca/internal/aluno"
)

func (a *App) listAlunos() ([]aluno.Aluno, error) {
	alunos, err := a.Alunos.List()
	if err != nil {
		return nil, err
	}
	if len(alunos) == 0 {
		fmt.Fprintln(a.out, "Nenhum aluno cadastrado.")
		return nil, nil
	}
	rows := make([][]string, len(alunos))
	for i, al := range alunos {
		rows[i] = []string{strconv.Itoa(al.Matricula), al.Nome, al.Turma}
	}
	a.table([]string{"MATRÍCULA", "NOME", "TURMA"}, rows)
	return alunos, nil
}

func (a *App) insertAluno() error {
	al := aluno.Aluno{
		Matricula: a.in.Int("Matrícula"),
		Nome:      a.in.Text("Nome"),
		Email:     a.in.Text("E-mail"),
		Turma:     a.in.Text("Turma"),
	}
	created, err := a.Alunos.Create(al)
	if err != nil {
		return err
	}
	a.success("Aluno cadastrado:\n  " + created.String())
	return nil
}

func (a *App) updateAluno() error {
	alunos, err := a.listAlunos()
	if err != nil || len(alunos) == 0 {
		return err
	}
	al, err := a.Alunos.Get(a.in.Int("\nMatrícula do aluno a atualizar"))
	if err != nil {
		return err
	}

	fmt.Fprintln(a.out, "\nInforme os novos valores (Enter mantém o valor atual):")
	al.Nome = a.in.TextOr("Nome", al.Nome)
	al.Email = a.in.TextOr("E-mail", al.Email)
	al.Turma = a.in.TextOr("Turma", al.Turma)

	updated, err := a.Alunos.Update(al)
	if err != nil {
		return err
	}
	a.success("Registro atualizado:\n  " + updated.String())
	return nil
}

func (a *App) removeAluno() error {
	alunos, err := a.listAlunos()
	if err != nil || len(alunos) == 0 {
		return err
	}
	al, err := a.Alunos.Get(a.in.Int("\nMatrícula do aluno a remover"))
	if err != nil {
		return err
	}
	fmt.Fprintln(a.out, "\n  "+al.String())
	return a.removeWithDependents(a.Alunos, al.Matricula)
}
