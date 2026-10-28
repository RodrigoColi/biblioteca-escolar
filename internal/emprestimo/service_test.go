package emprestimo

import (
	"errors"
	"testing"

	"biblioteca/internal/aluno"
	"biblioteca/internal/livro"
)

type fakeRepo struct {
	emprestimos map[int]Emprestimo
	nextID      int
	err         error
}

func newFakeRepo(emprestimos ...Emprestimo) *fakeRepo {
	r := &fakeRepo{emprestimos: map[int]Emprestimo{}, nextID: 1}
	for _, e := range emprestimos {
		r.emprestimos[e.ID] = e
		r.nextID = max(r.nextID, e.ID+1)
	}
	return r
}

func (r *fakeRepo) Count() (int, error) { return len(r.emprestimos), r.err }

func (r *fakeRepo) Exists(id int) (bool, error) {
	_, ok := r.emprestimos[id]
	return ok, r.err
}

func (r *fakeRepo) FindAll() ([]Emprestimo, error) {
	var all []Emprestimo
	for _, e := range r.emprestimos {
		all = append(all, e)
	}
	return all, r.err
}

func (r *fakeRepo) FindByID(id int) (Emprestimo, error) {
	e, ok := r.emprestimos[id]
	if !ok {
		return e, ErrNotFound
	}
	return e, r.err
}

func (r *fakeRepo) Insert(e Emprestimo) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	e.ID = r.nextID
	r.nextID++
	r.emprestimos[e.ID] = e
	return e.ID, nil
}

func (r *fakeRepo) Update(e Emprestimo) error {
	if r.err == nil {
		r.emprestimos[e.ID] = e
	}
	return r.err
}

func (r *fakeRepo) Delete(id int) error {
	delete(r.emprestimos, id)
	return r.err
}

func (r *fakeRepo) CountOpenByLivro(idLivro, ignoreID int) (int, error) {
	n := 0
	for _, e := range r.emprestimos {
		if e.Livro.ID == idLivro && e.DataDevolucao == nil && e.ID != ignoreID {
			n++
		}
	}
	return n, r.err
}

type alunoFinder map[int]aluno.Aluno

func (f alunoFinder) Get(m int) (aluno.Aluno, error) {
	a, ok := f[m]
	if !ok {
		return a, aluno.ErrNotFound
	}
	return a, nil
}

type livroFinder map[int]livro.Livro

func (f livroFinder) Get(id int) (livro.Livro, error) {
	l, ok := f[id]
	if !ok {
		return l, livro.ErrNotFound
	}
	return l, nil
}

var (
	ana       = aluno.Aluno{Matricula: 1001, Nome: "Ana", Email: "ana@escola.com", Turma: "9A"}
	bruno     = aluno.Aluno{Matricula: 1002, Nome: "Bruno", Email: "bruno@escola.com", Turma: "9A"}
	casmurro  = livro.Livro{ID: 1, Titulo: "Dom Casmurro", Autor: "Machado de Assis", Genero: "Romance", AnoPublicacao: 1899}
	cortico   = livro.Livro{ID: 2, Titulo: "O Cortiço", Autor: "Aluísio Azevedo", Genero: "Romance", AnoPublicacao: 1890}
	alunos    = alunoFinder{ana.Matricula: ana, bruno.Matricula: bruno}
	livros    = livroFinder{casmurro.ID: casmurro, cortico.ID: cortico}
	inicio    = day(2026, 10, 1)
	previsto  = day(2026, 10, 15)
	devolvido = day(2026, 10, 10)
)

func newLoan(matricula, idLivro int) Emprestimo {
	e := Emprestimo{DataEmprestimo: inicio, DataPrevista: previsto}
	e.Aluno.Matricula = matricula
	e.Livro.ID = idLivro
	return e
}

func newService(repo *fakeRepo) *Service {
	return NewService(repo, alunos, livros)
}

func TestCreateLoadsAssociations(t *testing.T) {
	repo := newFakeRepo()
	got, err := newService(repo).Create(newLoan(1001, 1))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 || got.Aluno != ana || got.Livro != casmurro {
		t.Errorf("got %+v", got)
	}
}

func TestCreateRules(t *testing.T) {
	before := day(2026, 9, 30)
	tests := []struct {
		name    string
		loan    func() Emprestimo
		wantErr error
		ok      bool
	}{
		{"aluno inexistente", func() Emprestimo { return newLoan(9999, 1) }, aluno.ErrNotFound, false},
		{"livro inexistente", func() Emprestimo { return newLoan(1001, 99) }, livro.ErrNotFound, false},
		{"previsão antes do empréstimo", func() Emprestimo {
			e := newLoan(1001, 1)
			e.DataPrevista = before
			return e
		}, nil, false},
		{"devolução antes do empréstimo", func() Emprestimo {
			e := newLoan(1001, 1)
			e.DataDevolucao = &before
			return e
		}, nil, false},
		{"datas iguais", func() Emprestimo {
			e := newLoan(1001, 1)
			e.DataPrevista = inicio
			e.DataDevolucao = &inicio
			return e
		}, nil, true},
	}
	for _, tt := range tests {
		repo := newFakeRepo()
		_, err := newService(repo).Create(tt.loan())
		switch {
		case tt.ok && err != nil:
			t.Errorf("%s: unexpected error %v", tt.name, err)
		case !tt.ok && err == nil:
			t.Errorf("%s: expected error", tt.name)
		case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
			t.Errorf("%s: got %v, want %v", tt.name, err, tt.wantErr)
		}
		if !tt.ok && len(repo.emprestimos) != 0 {
			t.Errorf("%s: invalid loan was stored", tt.name)
		}
	}
}

func TestCreateBookAlreadyLent(t *testing.T) {
	open := newLoan(1001, 1)
	open.ID = 1
	repo := newFakeRepo(open)

	if _, err := newService(repo).Create(newLoan(1002, 1)); err == nil {
		t.Error("expected error for book already lent")
	}

	returned := newLoan(1002, 1)
	returned.DataDevolucao = &devolvido
	if _, err := newService(repo).Create(returned); err != nil {
		t.Errorf("an already returned loan should not conflict: %v", err)
	}
}

func TestCreateAfterReturn(t *testing.T) {
	old := newLoan(1001, 1)
	old.ID = 1
	old.DataDevolucao = &devolvido
	if _, err := newService(newFakeRepo(old)).Create(newLoan(1002, 1)); err != nil {
		t.Errorf("returned book should be available: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	open := newLoan(1001, 1)
	open.ID = 1
	other := newLoan(1002, 2)
	other.ID = 2

	t.Run("não conflita consigo mesmo", func(t *testing.T) {
		repo := newFakeRepo(open, other)
		e := open
		e.DataPrevista = day(2026, 10, 20)
		got, err := newService(repo).Update(e)
		if err != nil {
			t.Fatal(err)
		}
		if !got.DataPrevista.Equal(day(2026, 10, 20)) {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("troca para livro emprestado", func(t *testing.T) {
		repo := newFakeRepo(open, other)
		e := open
		e.Livro = livro.Livro{ID: 2}
		if _, err := newService(repo).Update(e); err == nil {
			t.Error("expected error")
		}
		if repo.emprestimos[1].Livro.ID != 1 {
			t.Error("invalid update was stored")
		}
	})

	t.Run("registrar e limpar devolução", func(t *testing.T) {
		repo := newFakeRepo(open)
		svc := newService(repo)
		e := open
		e.DataDevolucao = &devolvido
		got, err := svc.Update(e)
		if err != nil || got.DataDevolucao == nil {
			t.Fatalf("got %+v, %v", got, err)
		}
		got.DataDevolucao = nil
		if got, err = svc.Update(got); err != nil || got.DataDevolucao != nil {
			t.Errorf("got %+v, %v", got, err)
		}
	})

	t.Run("inexistente", func(t *testing.T) {
		e := open
		e.ID = 99
		if _, err := newService(newFakeRepo()).Update(e); !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("aluno inexistente", func(t *testing.T) {
		e := open
		e.Aluno = aluno.Aluno{Matricula: 9999}
		if _, err := newService(newFakeRepo(open)).Update(e); !errors.Is(err, aluno.ErrNotFound) {
			t.Errorf("got %v", err)
		}
	})
}

func TestDelete(t *testing.T) {
	open := newLoan(1001, 1)
	open.ID = 1
	repo := newFakeRepo(open)
	svc := newService(repo)

	if err := svc.Delete(99); !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v", err)
	}
	if err := svc.Delete(1); err != nil {
		t.Fatal(err)
	}
	if len(repo.emprestimos) != 0 {
		t.Error("loan not removed")
	}
}

func TestRepositoryErrorsPropagate(t *testing.T) {
	boom := errors.New("conexão perdida")
	repo := newFakeRepo()
	repo.err = boom
	svc := newService(repo)

	if _, err := svc.Create(newLoan(1001, 1)); !errors.Is(err, boom) {
		t.Errorf("Create: %v", err)
	}
	if err := svc.Delete(1); !errors.Is(err, boom) {
		t.Errorf("Delete: %v", err)
	}
	if _, err := svc.List(); !errors.Is(err, boom) {
		t.Errorf("List: %v", err)
	}
}
