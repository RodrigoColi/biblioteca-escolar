package livro

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("livro não encontrado")
	ErrHasEmprestimos = errors.New("o livro possui empréstimos vinculados")
)

type Livro struct {
	ID            int
	Titulo        string
	Autor         string
	Genero        string
	AnoPublicacao int
}

func (l Livro) String() string {
	return fmt.Sprintf("ID: %d | Título: %s | Autor: %s | Gênero: %s | Ano: %d", l.ID, l.Titulo, l.Autor, l.Genero, l.AnoPublicacao)
}
