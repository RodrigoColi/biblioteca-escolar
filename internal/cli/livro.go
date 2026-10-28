package cli

import (
	"fmt"
	"strconv"

	"biblioteca/internal/livro"
)

func (a *App) listLivros() ([]livro.Livro, error) {
	livros, err := a.Livros.List()
	if err != nil {
		return nil, err
	}
	if len(livros) == 0 {
		fmt.Fprintln(a.out, "Nenhum livro cadastrado.")
		return nil, nil
	}
	rows := make([][]string, len(livros))
	for i, l := range livros {
		rows[i] = []string{strconv.Itoa(l.ID), l.Titulo, l.Autor, l.Genero}
	}
	a.table([]string{"ID", "TÍTULO", "AUTOR", "GÊNERO"}, rows)
	return livros, nil
}

func (a *App) insertLivro() error {
	l := livro.Livro{
		Titulo:        a.in.Text("Título"),
		Autor:         a.in.Text("Autor"),
		Genero:        a.in.Text("Gênero"),
		AnoPublicacao: a.in.Int("Ano de publicação"),
	}
	created, err := a.Livros.Create(l)
	if err != nil {
		return err
	}
	a.success("Livro cadastrado:\n  " + created.String())
	return nil
}

func (a *App) updateLivro() error {
	livros, err := a.listLivros()
	if err != nil || len(livros) == 0 {
		return err
	}
	l, err := a.Livros.Get(a.in.Int("\nID do livro a atualizar"))
	if err != nil {
		return err
	}

	fmt.Fprintln(a.out, "\nInforme os novos valores (Enter mantém o valor atual):")
	l.Titulo = a.in.TextOr("Título", l.Titulo)
	l.Autor = a.in.TextOr("Autor", l.Autor)
	l.Genero = a.in.TextOr("Gênero", l.Genero)
	l.AnoPublicacao = a.in.IntOr("Ano de publicação", l.AnoPublicacao)

	updated, err := a.Livros.Update(l)
	if err != nil {
		return err
	}
	a.success("Registro atualizado:\n  " + updated.String())
	return nil
}

func (a *App) removeLivro() error {
	livros, err := a.listLivros()
	if err != nil || len(livros) == 0 {
		return err
	}
	l, err := a.Livros.Get(a.in.Int("\nID do livro a remover"))
	if err != nil {
		return err
	}
	fmt.Fprintln(a.out, "\n  "+l.String())
	return a.removeWithDependents(a.Livros, l.ID)
}
