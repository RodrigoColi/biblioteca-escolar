package aluno

import (
	"errors"
	"testing"
)

type fakeRepo struct {
	alunos      map[int]Aluno
	emprestimos map[int]int
	err         error
}

func newFakeRepo(alunos ...Aluno) *fakeRepo {
	r := &fakeRepo{alunos: map[int]Aluno{}, emprestimos: map[int]int{}}
	for _, a := range alunos {
		r.alunos[a.Matricula] = a
	}
	return r
}

func (r *fakeRepo) Count() (int, error) { return len(r.alunos), r.err }

func (r *fakeRepo) Exists(m int) (bool, error) {
	_, ok := r.alunos[m]
	return ok, r.err
}

func (r *fakeRepo) FindAll() ([]Aluno, error) {
	var all []Aluno
	for _, a := range r.alunos {
		all = append(all, a)
	}
	return all, r.err
}

func (r *fakeRepo) FindByID(m int) (Aluno, error) {
	if r.err != nil {
		return Aluno{}, r.err
	}
	a, ok := r.alunos[m]
	if !ok {
		return a, ErrNotFound
	}
	return a, nil
}

func (r *fakeRepo) Insert(a Aluno) error {
	if r.err == nil {
		r.alunos[a.Matricula] = a
	}
	return r.err
}

func (r *fakeRepo) Update(a Aluno) error {
	if r.err == nil {
		r.alunos[a.Matricula] = a
	}
	return r.err
}

func (r *fakeRepo) Delete(m int) error {
	if r.emprestimos[m] > 0 {
		return errors.New("fk violation")
	}
	delete(r.alunos, m)
	return r.err
}

func (r *fakeRepo) CountEmprestimos(m int) (int, error) { return r.emprestimos[m], r.err }

func (r *fakeRepo) DeleteWithEmprestimos(m int) error {
	delete(r.emprestimos, m)
	delete(r.alunos, m)
	return r.err
}

var valid = Aluno{Matricula: 1001, Nome: "Ana", Email: "ana@escola.com", Turma: "9A"}

func TestCreate(t *testing.T) {
	repo := newFakeRepo()
	got, err := NewService(repo).Create(valid)
	if err != nil {
		t.Fatal(err)
	}
	if got != valid || repo.alunos[1001] != valid {
		t.Errorf("got %+v, stored %+v", got, repo.alunos[1001])
	}
}

func TestCreateValidation(t *testing.T) {
	tests := map[string]func(*Aluno){
		"matrícula zero":     func(a *Aluno) { a.Matricula = 0 },
		"matrícula negativa": func(a *Aluno) { a.Matricula = -1 },
		"nome vazio":         func(a *Aluno) { a.Nome = "   " },
		"e-mail sem arroba":  func(a *Aluno) { a.Email = "ana.escola.com" },
		"e-mail vazio":       func(a *Aluno) { a.Email = "" },
		"turma vazia":        func(a *Aluno) { a.Turma = "" },
	}
	for name, mutate := range tests {
		a := valid
		mutate(&a)
		repo := newFakeRepo()
		if _, err := NewService(repo).Create(a); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
		if len(repo.alunos) != 0 {
			t.Errorf("%s: invalid aluno was stored", name)
		}
	}
}

func TestCreateDuplicate(t *testing.T) {
	repo := newFakeRepo(valid)
	other := valid
	other.Nome = "Outra"
	if _, err := NewService(repo).Create(other); !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("got %v, want ErrAlreadyExists", err)
	}
	if repo.alunos[1001].Nome != "Ana" {
		t.Error("existing aluno was overwritten")
	}
}

func TestUpdate(t *testing.T) {
	repo := newFakeRepo(valid)
	changed := valid
	changed.Email = "novo@escola.com"
	got, err := NewService(repo).Update(changed)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "novo@escola.com" {
		t.Errorf("got %+v", got)
	}
}

func TestUpdateNotFoundAndInvalid(t *testing.T) {
	svc := NewService(newFakeRepo(valid))

	missing := valid
	missing.Matricula = 9999
	if _, err := svc.Update(missing); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}

	invalid := valid
	invalid.Nome = ""
	if _, err := svc.Update(invalid); err == nil {
		t.Error("expected validation error")
	}
}

func TestGetNotFound(t *testing.T) {
	if _, err := NewService(newFakeRepo()).Get(1); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v", err)
	}
}

func TestDelete(t *testing.T) {
	t.Run("sem empréstimos", func(t *testing.T) {
		repo := newFakeRepo(valid)
		if err := NewService(repo).Delete(1001, false); err != nil {
			t.Fatal(err)
		}
		if _, ok := repo.alunos[1001]; ok {
			t.Error("aluno not removed")
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		if err := NewService(newFakeRepo()).Delete(1001, false); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("com empréstimos sem cascata", func(t *testing.T) {
		repo := newFakeRepo(valid)
		repo.emprestimos[1001] = 2
		svc := NewService(repo)
		if n, _ := svc.Dependents(1001); n != 2 {
			t.Errorf("Dependents = %d", n)
		}
		if err := svc.Delete(1001, false); !errors.Is(err, ErrHasEmprestimos) {
			t.Errorf("got %v, want ErrHasEmprestimos", err)
		}
		if _, ok := repo.alunos[1001]; !ok || repo.emprestimos[1001] != 2 {
			t.Error("nothing should have been removed")
		}
	})

	t.Run("com empréstimos e cascata", func(t *testing.T) {
		repo := newFakeRepo(valid)
		repo.emprestimos[1001] = 2
		if err := NewService(repo).Delete(1001, true); err != nil {
			t.Fatal(err)
		}
		if _, ok := repo.alunos[1001]; ok || repo.emprestimos[1001] != 0 {
			t.Error("aluno and emprestimos should have been removed")
		}
	})
}

func TestRepositoryErrorsPropagate(t *testing.T) {
	boom := errors.New("conexão perdida")
	repo := newFakeRepo(valid)
	repo.err = boom
	svc := NewService(repo)

	other := valid
	other.Matricula = 2
	if _, err := svc.Create(other); !errors.Is(err, boom) {
		t.Errorf("Create: %v", err)
	}
	if _, err := svc.Update(valid); !errors.Is(err, boom) {
		t.Errorf("Update: %v", err)
	}
	if err := svc.Delete(1001, false); !errors.Is(err, boom) {
		t.Errorf("Delete: %v", err)
	}
	if _, err := svc.List(); !errors.Is(err, boom) {
		t.Errorf("List: %v", err)
	}
	if _, err := svc.Count(); !errors.Is(err, boom) {
		t.Errorf("Count: %v", err)
	}
}
