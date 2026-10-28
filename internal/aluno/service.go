package aluno

import (
	"errors"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}

func (s *Service) List() ([]Aluno, error) {
	return s.repo.FindAll()
}

func (s *Service) Get(matricula int) (Aluno, error) {
	return s.repo.FindByID(matricula)
}

func (s *Service) Create(a Aluno) (Aluno, error) {
	if err := validate(a); err != nil {
		return a, err
	}
	exists, err := s.repo.Exists(a.Matricula)
	if err != nil {
		return a, err
	}
	if exists {
		return a, ErrAlreadyExists
	}
	if err := s.repo.Insert(a); err != nil {
		return a, err
	}
	return s.repo.FindByID(a.Matricula)
}

func (s *Service) Update(a Aluno) (Aluno, error) {
	if err := validate(a); err != nil {
		return a, err
	}
	if err := s.mustExist(a.Matricula); err != nil {
		return a, err
	}
	if err := s.repo.Update(a); err != nil {
		return a, err
	}
	return s.repo.FindByID(a.Matricula)
}

func (s *Service) Dependents(matricula int) (int, error) {
	return s.repo.CountEmprestimos(matricula)
}

func (s *Service) Delete(matricula int, removeEmprestimos bool) error {
	if err := s.mustExist(matricula); err != nil {
		return err
	}
	if removeEmprestimos {
		return s.repo.DeleteWithEmprestimos(matricula)
	}
	n, err := s.repo.CountEmprestimos(matricula)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrHasEmprestimos
	}
	return s.repo.Delete(matricula)
}

func (s *Service) mustExist(matricula int) error {
	exists, err := s.repo.Exists(matricula)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func validate(a Aluno) error {
	switch {
	case a.Matricula <= 0:
		return errors.New("a matrícula deve ser um número positivo")
	case strings.TrimSpace(a.Nome) == "":
		return errors.New("o nome é obrigatório")
	case !strings.Contains(a.Email, "@"):
		return errors.New("e-mail inválido")
	case strings.TrimSpace(a.Turma) == "":
		return errors.New("a turma é obrigatória")
	}
	return nil
}
