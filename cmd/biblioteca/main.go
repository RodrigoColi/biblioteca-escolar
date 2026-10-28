package main

import (
	"log"
	"os"

	"biblioteca/internal/aluno"
	"biblioteca/internal/cli"
	"biblioteca/internal/config"
	"biblioteca/internal/database"
	"biblioteca/internal/emprestimo"
	"biblioteca/internal/input"
	"biblioteca/internal/livro"
	"biblioteca/internal/relatorio"
)

func main() {
	cfg := config.Load()

	if err := database.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	alunos := aluno.NewService(aluno.NewPostgres(db))
	livros := livro.NewService(livro.NewPostgres(db))

	app := cli.New(input.NewReader(os.Stdin, os.Stdout), os.Stdout, cli.Services{
		Alunos:      alunos,
		Livros:      livros,
		Emprestimos: emprestimo.NewService(emprestimo.NewPostgres(db), alunos, livros),
		Relatorios:  relatorio.NewService(relatorio.NewPostgres(db)),
	})
	app.Run()
}
