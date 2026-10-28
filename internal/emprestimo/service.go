package emprestimo

import (
	"errors"

	"biblioteca/internal/aluno"
	"biblioteca/internal/livro"
)

type AlunoFinder interface {
	Get(matricula int) (aluno.Aluno, error)
}

type LivroFinder interface {
	Get(id int) (livro.Livro, error)
}

type Service struct {
	repo   Repository
	alunos AlunoFinder
	livros LivroFinder
}

func NewService(repo Repository, alunos AlunoFinder, livros LivroFinder) *Service {
	return &Service{repo: repo, alunos: alunos, livros: livros}
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}

func (s *Service) List() ([]Emprestimo, error) {
	return s.repo.FindAll()
}

func (s *Service) Get(id int) (Emprestimo, error) {
	return s.repo.FindByID(id)
}

func (s *Service) Create(e Emprestimo) (Emprestimo, error) {
	if err := s.prepare(&e); err != nil {
		return e, err
	}
	id, err := s.repo.Insert(e)
	if err != nil {
		return e, err
	}
	return s.repo.FindByID(id)
}

func (s *Service) Update(e Emprestimo) (Emprestimo, error) {
	exists, err := s.repo.Exists(e.ID)
	if err != nil {
		return e, err
	}
	if !exists {
		return e, ErrNotFound
	}
	if err := s.prepare(&e); err != nil {
		return e, err
	}
	if err := s.repo.Update(e); err != nil {
		return e, err
	}
	return s.repo.FindByID(e.ID)
}

func (s *Service) Delete(id int) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return s.repo.Delete(id)
}

// prepare loads the associated aluno and livro and validates the loan rules.
func (s *Service) prepare(e *Emprestimo) error {
	a, err := s.alunos.Get(e.Aluno.Matricula)
	if err != nil {
		return err
	}
	l, err := s.livros.Get(e.Livro.ID)
	if err != nil {
		return err
	}
	e.Aluno, e.Livro = a, l

	if e.DataPrevista.Before(e.DataEmprestimo) {
		return errors.New("a data prevista não pode ser anterior à data do empréstimo")
	}
	if e.DataDevolucao != nil && e.DataDevolucao.Before(e.DataEmprestimo) {
		return errors.New("a data de devolução não pode ser anterior à data do empréstimo")
	}
	if e.DataDevolucao == nil {
		open, err := s.repo.CountOpenByLivro(e.Livro.ID, e.ID)
		if err != nil {
			return err
		}
		if open > 0 {
			return errors.New("este livro já está emprestado e ainda não foi devolvido")
		}
	}
	return nil
}
