package relatorio

type Repository interface {
	EmprestimosPorGenero() ([]PorGenero, error)
	EmprestimosDetalhados() ([]Detalhado, error)
}
