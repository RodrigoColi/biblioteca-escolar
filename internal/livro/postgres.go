package livro

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

const selectLivro = "select id_livro, titulo, autor, genero, ano_publicacao from livros"

func (r *Postgres) Count() (int, error) {
	return sqlutil.Count(r.db, "select count(1) total_livros from livros")
}

func (r *Postgres) Exists(id int) (bool, error) {
	n, err := sqlutil.Count(r.db, "select count(1) from livros where id_livro = "+sqlutil.Num(id))
	return n > 0, err
}

func (r *Postgres) FindAll() ([]Livro, error) {
	rows, err := r.db.Query(selectLivro + " order by id_livro")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var livros []Livro
	for rows.Next() {
		var l Livro
		if err := rows.Scan(&l.ID, &l.Titulo, &l.Autor, &l.Genero, &l.AnoPublicacao); err != nil {
			return nil, err
		}
		livros = append(livros, l)
	}
	return livros, rows.Err()
}

func (r *Postgres) FindByID(id int) (Livro, error) {
	var l Livro
	err := r.db.QueryRow(selectLivro+" where id_livro = "+sqlutil.Num(id)).
		Scan(&l.ID, &l.Titulo, &l.Autor, &l.Genero, &l.AnoPublicacao)
	if errors.Is(err, sql.ErrNoRows) {
		return l, ErrNotFound
	}
	return l, err
}

func (r *Postgres) Insert(l Livro) (int, error) {
	var id int
	err := r.db.QueryRow("insert into livros (titulo, autor, genero, ano_publicacao) values (" +
		sqlutil.Text(l.Titulo) + ", " + sqlutil.Text(l.Autor) + ", " +
		sqlutil.Text(l.Genero) + ", " + sqlutil.Num(l.AnoPublicacao) +
		") returning id_livro").Scan(&id)
	return id, err
}

func (r *Postgres) Update(l Livro) error {
	_, err := r.db.Exec("update livros set titulo = " + sqlutil.Text(l.Titulo) +
		", autor = " + sqlutil.Text(l.Autor) +
		", genero = " + sqlutil.Text(l.Genero) +
		", ano_publicacao = " + sqlutil.Num(l.AnoPublicacao) +
		" where id_livro = " + sqlutil.Num(l.ID))
	return err
}

func (r *Postgres) Delete(id int) error {
	_, err := r.db.Exec("delete from livros where id_livro = " + sqlutil.Num(id))
	return err
}

func (r *Postgres) CountEmprestimos(id int) (int, error) {
	return sqlutil.Count(r.db, "select count(1) from emprestimos where id_livro = "+sqlutil.Num(id))
}

func (r *Postgres) DeleteWithEmprestimos(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("delete from emprestimos where id_livro = " + sqlutil.Num(id)); err != nil {
		return err
	}
	if _, err := tx.Exec("delete from livros where id_livro = " + sqlutil.Num(id)); err != nil {
		return err
	}
	return tx.Commit()
}
