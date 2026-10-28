package aluno

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

const selectAluno = "select matricula, nome, email, turma from alunos"

func (r *Postgres) Count() (int, error) {
	return sqlutil.Count(r.db, "select count(1) total_alunos from alunos")
}

func (r *Postgres) Exists(matricula int) (bool, error) {
	n, err := sqlutil.Count(r.db, "select count(1) from alunos where matricula = "+sqlutil.Num(matricula))
	return n > 0, err
}

func (r *Postgres) FindAll() ([]Aluno, error) {
	rows, err := r.db.Query(selectAluno + " order by nome")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alunos []Aluno
	for rows.Next() {
		var a Aluno
		if err := rows.Scan(&a.Matricula, &a.Nome, &a.Email, &a.Turma); err != nil {
			return nil, err
		}
		alunos = append(alunos, a)
	}
	return alunos, rows.Err()
}

func (r *Postgres) FindByID(matricula int) (Aluno, error) {
	var a Aluno
	err := r.db.QueryRow(selectAluno+" where matricula = "+sqlutil.Num(matricula)).
		Scan(&a.Matricula, &a.Nome, &a.Email, &a.Turma)
	if errors.Is(err, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}

func (r *Postgres) Insert(a Aluno) error {
	_, err := r.db.Exec("insert into alunos (matricula, nome, email, turma) values (" +
		sqlutil.Num(a.Matricula) + ", " + sqlutil.Text(a.Nome) + ", " +
		sqlutil.Text(a.Email) + ", " + sqlutil.Text(a.Turma) + ")")
	return err
}

func (r *Postgres) Update(a Aluno) error {
	_, err := r.db.Exec("update alunos set nome = " + sqlutil.Text(a.Nome) +
		", email = " + sqlutil.Text(a.Email) +
		", turma = " + sqlutil.Text(a.Turma) +
		" where matricula = " + sqlutil.Num(a.Matricula))
	return err
}

func (r *Postgres) Delete(matricula int) error {
	_, err := r.db.Exec("delete from alunos where matricula = " + sqlutil.Num(matricula))
	return err
}

func (r *Postgres) CountEmprestimos(matricula int) (int, error) {
	return sqlutil.Count(r.db, "select count(1) from emprestimos where matricula = "+sqlutil.Num(matricula))
}

func (r *Postgres) DeleteWithEmprestimos(matricula int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("delete from emprestimos where matricula = " + sqlutil.Num(matricula)); err != nil {
		return err
	}
	if _, err := tx.Exec("delete from alunos where matricula = " + sqlutil.Num(matricula)); err != nil {
		return err
	}
	return tx.Commit()
}
