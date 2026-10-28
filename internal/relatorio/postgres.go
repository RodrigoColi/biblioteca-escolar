package relatorio

import (
	"database/sql"

	"biblioteca/internal/sqlutil"
)

var _ Repository = (*Postgres)(nil)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

const queryPorGenero = `select l.genero,
       count(e.id_emprestimo) as total,
       count(e.id_emprestimo) filter (where e.data_devolucao is null) as em_aberto,
       count(e.id_emprestimo) filter (where e.data_devolucao is null and e.data_prevista < current_date) as atrasados
  from livros l
  left join emprestimos e on e.id_livro = l.id_livro
 group by l.genero
 order by total desc, l.genero`

const queryDetalhado = `select e.id_emprestimo, a.nome, a.turma, l.titulo, l.autor,
       e.data_emprestimo, e.data_prevista, e.data_devolucao,
       case
         when e.data_devolucao is not null then 'Devolvido'
         when e.data_prevista < current_date then 'Atrasado'
         else 'Em aberto'
       end as situacao
  from emprestimos e
 inner join alunos a on a.matricula = e.matricula
 inner join livros l on l.id_livro = e.id_livro
 order by e.data_emprestimo desc`

func (r *Postgres) EmprestimosPorGenero() ([]PorGenero, error) {
	rows, err := r.db.Query(queryPorGenero)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PorGenero
	for rows.Next() {
		var g PorGenero
		if err := rows.Scan(&g.Genero, &g.Total, &g.EmAberto, &g.Atrasados); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	return result, rows.Err()
}

func (r *Postgres) EmprestimosDetalhados() ([]Detalhado, error) {
	rows, err := r.db.Query(queryDetalhado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Detalhado
	for rows.Next() {
		var d Detalhado
		var devolucao sql.NullTime
		if err := rows.Scan(&d.ID, &d.Aluno, &d.Turma, &d.Livro, &d.Autor,
			&d.DataEmprestimo, &d.DataPrevista, &devolucao, &d.Situacao); err != nil {
			return nil, err
		}
		d.DataDevolucao = sqlutil.TimePtr(devolucao)
		result = append(result, d)
	}
	return result, rows.Err()
}
