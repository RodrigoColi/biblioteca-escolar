package aluno

type Repository interface {
	Count() (int, error)
	Exists(matricula int) (bool, error)
	FindAll() ([]Aluno, error)
	FindByID(matricula int) (Aluno, error)
	Insert(a Aluno) error
	Update(a Aluno) error
	Delete(matricula int) error
	CountEmprestimos(matricula int) (int, error)
	DeleteWithEmprestimos(matricula int) error
}
