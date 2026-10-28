package emprestimo

import (
	"database/sql"
	"errors"

	"biblioteca/internal/sqlutil"
)

var _ Repository = (*Postgres)(nil)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

const selectEmprestimo = `select e.id_emprestimo, e.data_emprestimo, e.data_prevista, e.data_devolucao,
       a.matricula, a.nome, a.email, a.turma,
       l.id_livro, l.titulo, l.autor, l.genero, l.ano_publicacao
  from emprestimos e
  join alunos a on a.matricula = e.matricula
  join livros l on l.id_livro = e.id_livro`

type scanner interface {
	Scan(dest ...any) error
}

func scan(s scanner) (Emprestimo, error) {
	var e Emprestimo
	var devolucao sql.NullTime
	err := s.Scan(&e.ID, &e.DataEmprestimo, &e.DataPrevista, &devolucao,
		&e.Aluno.Matricula, &e.Aluno.Nome, &e.Aluno.Email, &e.Aluno.Turma,
		&e.Livro.ID, &e.Livro.Titulo, &e.Livro.Autor, &e.Livro.Genero, &e.Livro.AnoPublicacao)
	e.DataDevolucao = sqlutil.TimePtr(devolucao)
	return e, err
}

func (r *Postgres) Count() (int, error) {
	return sqlutil.Count(r.db, "select count(1) total_emprestimos from emprestimos")
}

func (r *Postgres) Exists(id int) (bool, error) {
	n, err := sqlutil.Count(r.db, "select count(1) from emprestimos where id_emprestimo = "+sqlutil.Num(id))
	return n > 0, err
}

func (r *Postgres) CountOpenByLivro(idLivro, ignoreID int) (int, error) {
	return sqlutil.Count(r.db, "select count(1) from emprestimos where data_devolucao is null and id_livro = "+
		sqlutil.Num(idLivro)+" and id_emprestimo <> "+sqlutil.Num(ignoreID))
}

func (r *Postgres) FindAll() ([]Emprestimo, error) {
	rows, err := r.db.Query(selectEmprestimo + " order by e.id_emprestimo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emprestimos []Emprestimo
	for rows.Next() {
		e, err := scan(rows)
		if err != nil {
			return nil, err
		}
		emprestimos = append(emprestimos, e)
	}
	return emprestimos, rows.Err()
}

func (r *Postgres) FindByID(id int) (Emprestimo, error) {
	e, err := scan(r.db.QueryRow(selectEmprestimo + " where e.id_emprestimo = " + sqlutil.Num(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (r *Postgres) Insert(e Emprestimo) (int, error) {
	var id int
	err := r.db.QueryRow("insert into emprestimos (matricula, id_livro, data_emprestimo, data_prevista, data_devolucao) values (" +
		sqlutil.Num(e.Aluno.Matricula) + ", " + sqlutil.Num(e.Livro.ID) + ", " +
		sqlutil.Date(e.DataEmprestimo) + ", " + sqlutil.Date(e.DataPrevista) + ", " +
		sqlutil.NullableDate(e.DataDevolucao) + ") returning id_emprestimo").Scan(&id)
	return id, err
}

func (r *Postgres) Update(e Emprestimo) error {
	_, err := r.db.Exec("update emprestimos set matricula = " + sqlutil.Num(e.Aluno.Matricula) +
		", id_livro = " + sqlutil.Num(e.Livro.ID) +
		", data_emprestimo = " + sqlutil.Date(e.DataEmprestimo) +
		", data_prevista = " + sqlutil.Date(e.DataPrevista) +
		", data_devolucao = " + sqlutil.NullableDate(e.DataDevolucao) +
		" where id_emprestimo = " + sqlutil.Num(e.ID))
	return err
}

func (r *Postgres) Delete(id int) error {
	_, err := r.db.Exec("delete from emprestimos where id_emprestimo = " + sqlutil.Num(id))
	return err
}
