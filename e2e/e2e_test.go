// Package e2e builds the real binary and drives it through stdin, checking the
// console output and the resulting database state.
package e2e

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"biblioteca/internal/database"
	"biblioteca/internal/testdb"
)

const (
	enter = ""
	sim   = "S"
	nao   = "N"

	relatorios = "1"
	inserir    = "2"
	remover    = "3"
	atualizar  = "4"
	sair       = "5"

	alunos      = "1"
	livros      = "2"
	emprestimos = "3"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "biblioteca-e2e")
	if err != nil {
		panic(err)
	}
	binPath = filepath.Join(dir, "biblioteca")
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func binary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		out, err := exec.Command("go", "build", "-o", binPath, "../cmd/biblioteca").CombinedOutput()
		if err != nil {
			buildErr = errors.New(string(out))
		}
	})
	if buildErr != nil {
		t.Fatalf("build: %v", buildErr)
	}
	return binPath
}

type result struct {
	out  string
	code int
}

func run(t *testing.T, dbURL string, lines ...string) result {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary(t))
	cmd.Env = append(os.Environ(), "DATABASE_URL="+dbURL)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n") + "\n")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("timeout; output:\n%s", out)
	}

	code := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return result{out: strings.ReplaceAll(string(out), "\x1b[H\x1b[2J", ""), code: code}
}

func setup(t *testing.T) (string, *sql.DB) {
	t.Helper()
	url := testdb.URL(t, "test_e2e")
	db, err := database.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return url, db
}

func (r result) contains(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(r.out, s) {
			t.Errorf("output missing %q\n--- output ---\n%s", s, r.out)
		}
	}
}

func (r result) notContains(t *testing.T, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if strings.Contains(r.out, s) {
			t.Errorf("output should not contain %q", s)
		}
	}
}

func (r result) matches(t *testing.T, patterns ...string) {
	t.Helper()
	for _, p := range patterns {
		if !regexp.MustCompile(p).MatchString(r.out) {
			t.Errorf("output does not match %q\n--- output ---\n%s", p, r.out)
		}
	}
}

func (r result) count(t *testing.T, sub string, want int) {
	t.Helper()
	if n := strings.Count(r.out, sub); n != want {
		t.Errorf("%q appears %d times, want %d", sub, n, want)
	}
}

func (r result) exitOK(t *testing.T) {
	t.Helper()
	if r.code != 0 {
		t.Fatalf("exit code %d\n%s", r.code, r.out)
	}
}

func queryString(t *testing.T, db *sql.DB, query string) string {
	t.Helper()
	var s sql.NullString
	if err := db.QueryRow(query).Scan(&s); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return s.String
}

func TestSplashScreen(t *testing.T) {
	url, _ := setup(t)
	r := run(t, url, enter, sair)
	r.exitOK(t)
	r.contains(t, "SISTEMA DE BIBLIOTECA ESCOLAR", "CRIADO POR:", "Rodrigo Leite Colli Pedroso",
		"Howard Roatti", "Banco de Dados", "2026/2", "MENU PRINCIPAL", "Até logo!")
	r.matches(t, `ALUNOS:\s+5`, `LIVROS:\s+6`, `EMPRÉSTIMOS:\s+6`)
}

func TestMainMenuInvalidOptions(t *testing.T) {
	url, _ := setup(t)
	r := run(t, url, enter, "0", "9", "abc", inserir, "0", sair)
	r.exitOK(t)
	r.count(t, "Escolha uma opção entre 1 e 5", 3)
	r.count(t, "MENU PRINCIPAL", 2)
}

func TestRelatorios(t *testing.T) {
	url, _ := setup(t)
	r := run(t, url, enter, relatorios, "7", "1", enter, "2", enter, "0", sair)
	r.exitOK(t)
	r.contains(t, "Escolha uma opção entre 0 e 2", "EMPRÉSTIMOS POR GÊNERO", "EMPRÉSTIMOS DETALHADOS")
	r.matches(t,
		`GÊNERO\s+TOTAL\s+EM ABERTO\s+ATRASADOS`,
		`Romance\s+2\s+1\s+1`,
		`Ficção\s+1\s+1\s+1`,
		`Ana Souza\s+9A\s+Dom Casmurro\s+Machado de Assis\s+01/09/2026\s+15/09/2026\s+14/09/2026\s+Devolvido`,
		`Carla Mendes\s+8B\s+A Revolução dos Bichos.*Atrasado`,
	)
}

func TestInserirAlunoComEntradasInvalidasELoop(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, inserir, alunos,
		"abc", "1006", enter, "Fulano", "sem-arroba", "6A",
		sim, alunos, "1006", "Fulano O'Neil", "fulano@escola.com", "6A",
		nao, sair)
	r.exitOK(t)
	r.contains(t, "Informe um número inteiro", "Campo obrigatório", "Erro: e-mail inválido",
		"Aluno cadastrado", "Matrícula: 1006 | Nome: Fulano O'Neil")
	r.count(t, "Escolha a entidade", 2)
	r.count(t, "Deseja inserir mais algum registro?", 2)
	r.count(t, "MENU PRINCIPAL", 2)

	if got := queryString(t, db, "select nome from alunos where matricula = 1006"); got != "Fulano O'Neil" {
		t.Errorf("nome = %q", got)
	}
}

func TestInserirMatriculaDuplicada(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, inserir, alunos, "1001", "Outro", "outro@escola.com", "1A", nao, sair)
	r.exitOK(t)
	r.contains(t, "já existe um aluno com essa matrícula")
	if got := queryString(t, db, "select nome from alunos where matricula = 1001"); got != "Ana Souza" {
		t.Errorf("nome = %q", got)
	}
}

func TestInserirLivro(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, inserir, livros,
		"Clean Code", "Robert Martin", "Técnico", "dois mil", "3000",
		sim, livros, "Clean Code", "Robert Martin", "Técnico", "2008",
		nao, sair)
	r.exitOK(t)
	r.contains(t, "Informe um número inteiro", "ano de publicação inválido", "Livro cadastrado", "ID: 7 | Título: Clean Code")
	if n := testdb.Count(t, db, "select count(1) from livros"); n != 7 {
		t.Errorf("livros = %d, want 7", n)
	}
}

func TestInserirEmprestimoRegras(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, inserir, emprestimos,
		"9999", "1", enter, enter, enter,
		sim, emprestimos, "1002", "4", enter, enter, enter,
		sim, emprestimos, "1002", "1", "10/10/2026", "01/10/2026", enter,
		sim, emprestimos, "1002", "1", "32/10/2026", "10/10/2026", "24/10/2026", "20/10/2026",
		nao, sair)
	r.exitOK(t)
	r.contains(t,
		"Erro: aluno não encontrado",
		"Erro: este livro já está emprestado",
		"Erro: a data prevista não pode ser anterior",
		"Data inválida",
		"Empréstimo cadastrado",
		"Aluno: Bruno Lima (1002) | Livro: Dom Casmurro (1)",
	)
	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 7 {
		t.Errorf("emprestimos = %d, want 7", n)
	}
	if got := queryString(t, db, "select to_char(data_devolucao, 'DD/MM/YYYY') from emprestimos where id_emprestimo = 7"); got != "20/10/2026" {
		t.Errorf("data_devolucao = %q", got)
	}
}

func TestRemoverAlunoComEmprestimosRecusandoCascata(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, remover, alunos, "1001", sim, nao, nao, sair)
	r.exitOK(t)
	r.contains(t, "MATRÍCULA", "Ana Souza", "referenciado por 2 empréstimo(s)", "não pode ser removido")
	r.notContains(t, "Registro removido")
	if n := testdb.Count(t, db, "select count(1) from emprestimos where matricula = 1001"); n != 2 {
		t.Errorf("emprestimos of 1001 = %d, want 2", n)
	}
	if n := testdb.Count(t, db, "select count(1) from alunos where matricula = 1001"); n != 1 {
		t.Error("aluno should not have been removed")
	}
}

func TestRemoverAlunoComCascata(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, remover, alunos, "1001", sim, sim, nao, sair)
	r.exitOK(t)
	r.contains(t, "Registro removido com sucesso")
	if n := testdb.Count(t, db, "select count(1) from alunos"); n != 4 {
		t.Errorf("alunos = %d, want 4", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 4 {
		t.Errorf("emprestimos = %d, want 4", n)
	}
}

func TestRemoverLivroComCascataEEmprestimo(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, remover,
		livros, "4", sim, sim,
		sim, emprestimos, "1", sim,
		nao, sair)
	r.exitOK(t)
	r.count(t, "Registro removido com sucesso", 2)
	if n := testdb.Count(t, db, "select count(1) from livros"); n != 5 {
		t.Errorf("livros = %d, want 5", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos where id_emprestimo in (1, 3)"); n != 0 {
		t.Errorf("loans 1 and 3 should be gone, %d left", n)
	}
	if n := testdb.Count(t, db, "select count(1) from emprestimos"); n != 4 {
		t.Errorf("emprestimos = %d, want 4", n)
	}
}

func TestRemoverCanceladoVoltaAoMenuPrincipal(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, remover, livros, "6", nao, enter, sair)
	r.exitOK(t)
	r.contains(t, "Operação cancelada. Voltando ao menu principal.")
	r.notContains(t, "Deseja remover mais algum registro?")
	r.count(t, "MENU PRINCIPAL", 2)
	if n := testdb.Count(t, db, "select count(1) from livros where id_livro = 6"); n != 1 {
		t.Error("livro should not have been removed")
	}
}

func TestRegistroInexistente(t *testing.T) {
	url, _ := setup(t)
	r := run(t, url, enter,
		remover, livros, "999", nao,
		atualizar, alunos, "9999", nao,
		remover, emprestimos, "999", nao,
		sair)
	r.exitOK(t)
	r.contains(t, "Erro: livro não encontrado", "Erro: aluno não encontrado", "Erro: empréstimo não encontrado")
	r.notContains(t, "Confirma a remoção")
}

func TestAtualizarAlunoMantendoValores(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, atualizar, alunos, "1002", enter, "bruno.novo@escola.com", "9B", nao, sair)
	r.exitOK(t)
	r.contains(t, "Nome [Bruno Lima]", "Registro atualizado",
		"Matrícula: 1002 | Nome: Bruno Lima | E-mail: bruno.novo@escola.com | Turma: 9B")
	if got := queryString(t, db, "select nome || '|' || email || '|' || turma from alunos where matricula = 1002"); got != "Bruno Lima|bruno.novo@escola.com|9B" {
		t.Errorf("aluno = %q", got)
	}
}

func TestAtualizarLivroInvalidoNaoAltera(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, atualizar, livros, "1", enter, enter, enter, "3000", nao, sair)
	r.exitOK(t)
	r.contains(t, "Ano de publicação [1899]", "Erro: ano de publicação inválido")
	if n := testdb.Count(t, db, "select ano_publicacao from livros where id_livro = 1"); n != 1899 {
		t.Errorf("ano = %d, want 1899", n)
	}
}

func TestAtualizarEmprestimoDevolucao(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, atualizar,
		emprestimos, "3", enter, enter, enter, enter, "20/10/2026",
		sim, emprestimos, "3", enter, enter, enter, enter, "-",
		nao, sair)
	r.exitOK(t)
	r.contains(t, "Devolução: 20/10/2026 | Situação: Devolvido", "Devolução: - | Situação: Atrasado")
	r.count(t, "Registro atualizado", 2)
	if got := queryString(t, db, "select data_devolucao::text from emprestimos where id_emprestimo = 3"); got != "" {
		t.Errorf("data_devolucao = %q, want NULL", got)
	}
}

func TestNomeComAspasETentativaDeInjecao(t *testing.T) {
	url, db := setup(t)
	nome := "Robert'); DROP TABLE alunos;--"
	r := run(t, url, enter, inserir, alunos, "1010", nome, "r@escola.com", "1A", nao, sair)
	r.exitOK(t)
	r.contains(t, "Aluno cadastrado")
	if got := queryString(t, db, "select nome from alunos where matricula = 1010"); got != nome {
		t.Errorf("nome = %q", got)
	}
	if n := testdb.Count(t, db, "select count(1) from alunos"); n != 6 {
		t.Errorf("alunos = %d, want 6", n)
	}
}

func TestFimInesperadoDaEntrada(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, inserir, alunos, "1006")
	r.exitOK(t)
	r.contains(t, "Entrada encerrada.")
	if n := testdb.Count(t, db, "select count(1) from alunos where matricula = 1006"); n != 0 {
		t.Error("partial input should not create a record")
	}
}

func TestReinicioNaoReaplicaCargaInicial(t *testing.T) {
	url, _ := setup(t)
	run(t, url, enter, inserir, alunos, "1006", "Fulano", "f@escola.com", "6A", nao, sair).exitOK(t)

	r := run(t, url, enter, sair)
	r.exitOK(t)
	r.matches(t, `ALUNOS:\s+6`, `LIVROS:\s+6`, `EMPRÉSTIMOS:\s+6`)
}

func TestBancoIndisponivel(t *testing.T) {
	r := run(t, "postgres://x:y@127.0.0.1:1/x?sslmode=disable&connect_timeout=2", enter, sair)
	if r.code == 0 {
		t.Fatalf("expected non-zero exit code\n%s", r.out)
	}
	r.contains(t, "conectando ao banco")
}

func TestAtualizarEmprestimoEmAbertoNaoConflitaConsigoMesmo(t *testing.T) {
	url, db := setup(t)
	r := run(t, url, enter, atualizar, emprestimos, "4", enter, enter, enter, "10/11/2026", enter, nao, sair)
	r.exitOK(t)
	r.contains(t, "Registro atualizado", "Previsão: 10/11/2026")
	r.notContains(t, "já está emprestado")
	if got := queryString(t, db, "select to_char(data_prevista, 'DD/MM/YYYY') from emprestimos where id_emprestimo = 4"); got != "10/11/2026" {
		t.Errorf("data_prevista = %q", got)
	}
}
