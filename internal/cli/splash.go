package cli

import (
	"fmt"
	"strings"
)

var integrantes = []string{
	"Rodrigo Leite Colli Pedroso",
}

func (a *App) splash() {
	line := strings.Repeat("#", 64)
	fmt.Fprint(a.out, "\033[H\033[2J")
	fmt.Fprintln(a.out, line)
	fmt.Fprintln(a.out, "#               SISTEMA DE BIBLIOTECA ESCOLAR                  #")
	fmt.Fprintln(a.out, line)

	fmt.Fprintln(a.out, "\n TOTAL DE REGISTROS EXISTENTES:")
	counters := []struct {
		name  string
		count func() (int, error)
	}{
		{"ALUNOS", a.Alunos.Count},
		{"LIVROS", a.Livros.Count},
		{"EMPRÉSTIMOS", a.Emprestimos.Count},
	}
	for i, c := range counters {
		n, err := c.count()
		if err != nil {
			fmt.Fprintf(a.out, "   %d - %-12s erro: %v\n", i+1, c.name+":", err)
			continue
		}
		fmt.Fprintf(a.out, "   %d - %-12s %d\n", i+1, c.name+":", n)
	}

	fmt.Fprintln(a.out, "\n CRIADO POR:")
	for _, nome := range integrantes {
		fmt.Fprintln(a.out, "   "+nome)
	}
	fmt.Fprintln(a.out, "\n PROFESSOR:  Howard Roatti")
	fmt.Fprintln(a.out, " DISCIPLINA: Banco de Dados")
	fmt.Fprintln(a.out, " SEMESTRE:   2026/2")
	fmt.Fprintln(a.out, line)
}
