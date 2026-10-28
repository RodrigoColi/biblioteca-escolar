package emprestimo

type Repository interface {
	Count() (int, error)
	Exists(id int) (bool, error)
	FindAll() ([]Emprestimo, error)
	FindByID(id int) (Emprestimo, error)
	Insert(e Emprestimo) (int, error)
	Update(e Emprestimo) error
	Delete(id int) error
	CountOpenByLivro(idLivro, ignoreID int) (int, error)
}
