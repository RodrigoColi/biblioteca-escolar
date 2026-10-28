package livro

type Repository interface {
	Count() (int, error)
	Exists(id int) (bool, error)
	FindAll() ([]Livro, error)
	FindByID(id int) (Livro, error)
	Insert(l Livro) (int, error)
	Update(l Livro) error
	Delete(id int) error
	CountEmprestimos(id int) (int, error)
	DeleteWithEmprestimos(id int) error
}
