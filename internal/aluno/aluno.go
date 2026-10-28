package aluno

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("aluno não encontrado")
	ErrAlreadyExists  = errors.New("já existe um aluno com essa matrícula")
	ErrHasEmprestimos = errors.New("o aluno possui empréstimos vinculados")
)

type Aluno struct {
	Matricula int
	Nome      string
	Email     string
	Turma     string
}

func (a Aluno) String() string {
	return fmt.Sprintf("Matrícula: %d | Nome: %s | E-mail: %s | Turma: %s", a.Matricula, a.Nome, a.Email, a.Turma)
}
