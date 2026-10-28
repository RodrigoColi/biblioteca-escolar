package livro

import (
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	livros      map[int]Livro
	emprestimos map[int]int
	nextID      int
	err         error
}

func newFakeRepo(livros ...Livro) *fakeRepo {
	r := &fakeRepo{livros: map[int]Livro{}, emprestimos: map[int]int{}, nextID: 1}
	for _, l := range livros {
		r.livros[l.ID] = l
		r.nextID = max(r.nextID, l.ID+1)
	}
	return r
}

func (r *fakeRepo) Count() (int, error) { return len(r.livros), r.err }

func (r *fakeRepo) Exists(id int) (bool, error) {
	_, ok := r.livros[id]
	return ok, r.err
}

func (r *fakeRepo) FindAll() ([]Livro, error) {
	var all []Livro
	for _, l := range r.livros {
		all = append(all, l)
	}
	return all, r.err
}

func (r *fakeRepo) FindByID(id int) (Livro, error) {
	if r.err != nil {
		return Livro{}, r.err
	}
	l, ok := r.livros[id]
	if !ok {
		return l, ErrNotFound
	}
	return l, nil
}

func (r *fakeRepo) Insert(l Livro) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	l.ID = r.nextID
	r.nextID++
	r.livros[l.ID] = l
	return l.ID, nil
}

func (r *fakeRepo) Update(l Livro) error {
	if r.err == nil {
		r.livros[l.ID] = l
	}
	return r.err
}

func (r *fakeRepo) Delete(id int) error {
	if r.emprestimos[id] > 0 {
		return errors.New("fk violation")
	}
	delete(r.livros, id)
	return r.err
}

func (r *fakeRepo) CountEmprestimos(id int) (int, error) { return r.emprestimos[id], r.err }

func (r *fakeRepo) DeleteWithEmprestimos(id int) error {
	delete(r.emprestimos, id)
	delete(r.livros, id)
	return r.err
}

var valid = Livro{Titulo: "Dom Casmurro", Autor: "Machado de Assis", Genero: "Romance", AnoPublicacao: 1899}

func TestCreateAssignsID(t *testing.T) {
	repo := newFakeRepo(Livro{ID: 5, Titulo: "x", Autor: "y", Genero: "z", AnoPublicacao: 2000})
	got, err := NewService(repo).Create(valid)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 6 || got.Titulo != valid.Titulo {
		t.Errorf("got %+v", got)
	}
}

func TestCreateValidation(t *testing.T) {
	tests := map[string]func(*Livro){
		"título vazio":  func(l *Livro) { l.Titulo = " " },
		"autor vazio":   func(l *Livro) { l.Autor = "" },
		"gênero vazio":  func(l *Livro) { l.Genero = "" },
		"ano zero":      func(l *Livro) { l.AnoPublicacao = 0 },
		"ano negativo":  func(l *Livro) { l.AnoPublicacao = -10 },
		"ano no futuro": func(l *Livro) { l.AnoPublicacao = time.Now().Year() + 1 },
	}
	for name, mutate := range tests {
		l := valid
		mutate(&l)
		repo := newFakeRepo()
		if _, err := NewService(repo).Create(l); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
		if len(repo.livros) != 0 {
			t.Errorf("%s: invalid livro was stored", name)
		}
	}

	current := valid
	current.AnoPublicacao = time.Now().Year()
	if _, err := NewService(newFakeRepo()).Create(current); err != nil {
		t.Errorf("current year should be valid: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	existing := valid
	existing.ID = 1
	repo := newFakeRepo(existing)
	svc := NewService(repo)

	existing.Genero = "Clássico"
	got, err := svc.Update(existing)
	if err != nil || got.Genero != "Clássico" {
		t.Errorf("got %+v, %v", got, err)
	}

	missing := existing
	missing.ID = 99
	if _, err := svc.Update(missing); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}

	invalid := existing
	invalid.Titulo = ""
	if _, err := svc.Update(invalid); err == nil {
		t.Error("expected validation error")
	}
	if repo.livros[1].Titulo != valid.Titulo {
		t.Error("invalid update was stored")
	}
}

func TestDelete(t *testing.T) {
	existing := valid
	existing.ID = 1

	t.Run("inexistente", func(t *testing.T) {
		if err := NewService(newFakeRepo()).Delete(1, true); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("com empréstimos sem cascata", func(t *testing.T) {
		repo := newFakeRepo(existing)
		repo.emprestimos[1] = 1
		if err := NewService(repo).Delete(1, false); !errors.Is(err, ErrHasEmprestimos) {
			t.Errorf("got %v", err)
		}
		if _, ok := repo.livros[1]; !ok {
			t.Error("livro should not have been removed")
		}
	})

	t.Run("com empréstimos e cascata", func(t *testing.T) {
		repo := newFakeRepo(existing)
		repo.emprestimos[1] = 3
		if err := NewService(repo).Delete(1, true); err != nil {
			t.Fatal(err)
		}
		if _, ok := repo.livros[1]; ok || repo.emprestimos[1] != 0 {
			t.Error("livro and emprestimos should have been removed")
		}
	})

	t.Run("sem empréstimos", func(t *testing.T) {
		repo := newFakeRepo(existing)
		if err := NewService(repo).Delete(1, false); err != nil {
			t.Fatal(err)
		}
		if len(repo.livros) != 0 {
			t.Error("livro not removed")
		}
	})
}

func TestRepositoryErrorsPropagate(t *testing.T) {
	boom := errors.New("conexão perdida")
	repo := newFakeRepo()
	repo.err = boom
	svc := NewService(repo)

	if _, err := svc.Create(valid); !errors.Is(err, boom) {
		t.Errorf("Create: %v", err)
	}
	if _, err := svc.Get(1); !errors.Is(err, boom) {
		t.Errorf("Get: %v", err)
	}
	if err := svc.Delete(1, false); !errors.Is(err, boom) {
		t.Errorf("Delete: %v", err)
	}
}
