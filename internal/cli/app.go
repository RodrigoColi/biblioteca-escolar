package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"biblioteca/internal/aluno"
	"biblioteca/internal/emprestimo"
	"biblioteca/internal/input"
	"biblioteca/internal/livro"
	"biblioteca/internal/relatorio"
)

var errCanceled = errors.New("operação cancelada")

type Services struct {
	Alunos      *aluno.Service
	Livros      *livro.Service
	Emprestimos *emprestimo.Service
	Relatorios  *relatorio.Service
}

type App struct {
	in  *input.Reader
	out io.Writer
	Services
}

func New(in *input.Reader, out io.Writer, s Services) *App {
	return &App{in: in, out: out, Services: s}
}

type entity struct {
	name                   string
	insert, remove, update func() error
}

func (a *App) entities() []entity {
	return []entity{
		{"Alunos", a.insertAluno, a.removeAluno, a.updateAluno},
		{"Livros", a.insertLivro, a.removeLivro, a.updateLivro},
		{"Empréstimos", a.insertEmprestimo, a.removeEmprestimo, a.updateEmprestimo},
	}
}

func (a *App) Run() {
	a.splash()
	a.in.Pause()

	for {
		a.header("MENU PRINCIPAL")
		fmt.Fprintln(a.out, " 1 - Relatórios")
		fmt.Fprintln(a.out, " 2 - Inserir registros")
		fmt.Fprintln(a.out, " 3 - Remover registros")
		fmt.Fprintln(a.out, " 4 - Atualizar registros")
		fmt.Fprintln(a.out, " 5 - Sair")

		switch a.in.Option("\nEscolha uma opção", 1, 5) {
		case 1:
			a.reportsMenu()
		case 2:
			a.crudMenu("INSERIR REGISTROS", "inserir", func(e entity) func() error { return e.insert })
		case 3:
			a.crudMenu("REMOVER REGISTROS", "remover", func(e entity) func() error { return e.remove })
		case 4:
			a.crudMenu("ATUALIZAR REGISTROS", "atualizar", func(e entity) func() error { return e.update })
		case 5:
			fmt.Fprintln(a.out, "\nAté logo!")
			return
		}
	}
}

func (a *App) crudMenu(title, verb string, action func(entity) func() error) {
	for {
		a.header(title)
		entities := a.entities()
		for i, e := range entities {
			fmt.Fprintf(a.out, " %d - %s\n", i+1, e.name)
		}
		fmt.Fprintln(a.out, " 0 - Voltar ao menu principal")

		opt := a.in.Option("\nEscolha a entidade", 0, len(entities))
		if opt == 0 {
			return
		}

		a.header(title + " » " + strings.ToUpper(entities[opt-1].name))
		err := action(entities[opt-1])()
		if errors.Is(err, errCanceled) {
			fmt.Fprintln(a.out, "\nOperação cancelada. Voltando ao menu principal.")
			a.in.Pause()
			return
		}
		if err != nil {
			a.fail(err)
		}
		if !a.in.Confirm("\nDeseja " + verb + " mais algum registro?") {
			return
		}
	}
}

type removable interface {
	Dependents(id int) (int, error)
	Delete(id int, removeEmprestimos bool) error
}

func (a *App) removeWithDependents(svc removable, id int) error {
	if !a.in.Confirm("\nConfirma a remoção deste registro?") {
		return errCanceled
	}
	n, err := svc.Dependents(id)
	if err != nil {
		return err
	}
	removeEmprestimos := false
	if n > 0 {
		fmt.Fprintf(a.out, "\nAtenção: este registro é referenciado por %d empréstimo(s).\n", n)
		if !a.in.Confirm("Deseja remover também os empréstimos vinculados?") {
			fmt.Fprintln(a.out, "O registro não pode ser removido enquanto possuir empréstimos vinculados.")
			return nil
		}
		removeEmprestimos = true
	}
	if err := svc.Delete(id, removeEmprestimos); err != nil {
		return err
	}
	a.success("Registro removido com sucesso.")
	return nil
}

func (a *App) header(title string) {
	line := strings.Repeat("=", 64)
	fmt.Fprint(a.out, "\033[H\033[2J")
	fmt.Fprintf(a.out, "%s\n  BIBLIOTECA ESCOLAR » %s\n%s\n", line, title, line)
}

func (a *App) success(msg string) {
	fmt.Fprintln(a.out, "\n✔ "+msg)
}

func (a *App) fail(err error) {
	fmt.Fprintln(a.out, "\n✘ Erro: "+err.Error())
}

func (a *App) table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	seps := make([]string, len(headers))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", len([]rune(h)))
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Join(seps, "\t"))
	for _, r := range rows {
		fmt.Fprintln(w, strings.Join(r, "\t"))
	}
	w.Flush()
}

func today() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
