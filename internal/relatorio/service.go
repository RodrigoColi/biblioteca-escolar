package relatorio

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EmprestimosPorGenero() ([]PorGenero, error) {
	return s.repo.EmprestimosPorGenero()
}

func (s *Service) EmprestimosDetalhados() ([]Detalhado, error) {
	return s.repo.EmprestimosDetalhados()
}
