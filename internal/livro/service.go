package livro

import (
	"errors"
	"strings"
	"time"
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

func (s *Service) List() ([]Livro, error) {
	return s.repo.FindAll()
}

func (s *Service) Get(id int) (Livro, error) {
	return s.repo.FindByID(id)
}

func (s *Service) Create(l Livro) (Livro, error) {
	if err := validate(l); err != nil {
		return l, err
	}
	id, err := s.repo.Insert(l)
	if err != nil {
		return l, err
	}
	return s.repo.FindByID(id)
}

func (s *Service) Update(l Livro) (Livro, error) {
	if err := validate(l); err != nil {
		return l, err
	}
	if err := s.mustExist(l.ID); err != nil {
		return l, err
	}
	if err := s.repo.Update(l); err != nil {
		return l, err
	}
	return s.repo.FindByID(l.ID)
}

func (s *Service) Dependents(id int) (int, error) {
	return s.repo.CountEmprestimos(id)
}

func (s *Service) Delete(id int, removeEmprestimos bool) error {
	if err := s.mustExist(id); err != nil {
		return err
	}
	if removeEmprestimos {
		return s.repo.DeleteWithEmprestimos(id)
	}
	n, err := s.repo.CountEmprestimos(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrHasEmprestimos
	}
	return s.repo.Delete(id)
}

func (s *Service) mustExist(id int) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func validate(l Livro) error {
	switch {
	case strings.TrimSpace(l.Titulo) == "":
		return errors.New("o título é obrigatório")
	case strings.TrimSpace(l.Autor) == "":
		return errors.New("o autor é obrigatório")
	case strings.TrimSpace(l.Genero) == "":
		return errors.New("o gênero é obrigatório")
	case l.AnoPublicacao <= 0 || l.AnoPublicacao > time.Now().Year():
		return errors.New("ano de publicação inválido")
	}
	return nil
}
