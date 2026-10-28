package relatorio

import "time"

type PorGenero struct {
	Genero    string
	Total     int
	EmAberto  int
	Atrasados int
}

type Detalhado struct {
	ID             int
	Aluno          string
	Turma          string
	Livro          string
	Autor          string
	DataEmprestimo time.Time
	DataPrevista   time.Time
	DataDevolucao  *time.Time
	Situacao       string
}
