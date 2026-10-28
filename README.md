# Sistema de Biblioteca Escolar

Aplicação de console em Go para controle de **alunos**, **livros** e **empréstimos**, com relatórios de empréstimos por gênero e de empréstimos detalhados. Banco de dados: **PostgreSQL**.

Autor: Rodrigo Leite Colli Pedroso · Disciplina: Banco de Dados · Professor: Howard Roatti · Semestre 2026/2

Vídeo demonstrativo: _(link do YouTube a adicionar)_

## Estrutura

O código é organizado por funcionalidade. Cada pacote de domínio segue o mesmo formato:

| Arquivo         | Responsabilidade                                                        |
|-----------------|-------------------------------------------------------------------------|
| `<entidade>.go` | a entidade (struct + `String()`) e os erros do domínio                  |
| `repository.go` | interface do repositório (o que o serviço precisa do banco)             |
| `postgres.go`   | implementação da interface para PostgreSQL — SQL montado por concatenação, sem ORM |
| `service.go`    | regras de negócio: existência do registro, validações e dependências    |

```
cmd/biblioteca/        ponto de entrada (main): monta repositórios, serviços e interface
internal/aluno/        alunos
internal/livro/        livros
internal/emprestimo/   empréstimos (referencia um Aluno e um Livro)
internal/relatorio/    relatórios (GROUP BY e JOIN)
internal/sqlutil/      funções auxiliares para montar o SQL com segurança
internal/input/        leitura e validação da entrada do usuário no console
internal/cli/          splash screen, menus e telas
internal/config/       leitura de configuração (DATABASE_URL)
internal/database/     conexão e execução das migrations
migrations/            scripts de criação das tabelas e carga inicial (golang-migrate)
docs/                  diagrama relacional
```

## Modelo relacional

![Diagrama relacional](docs/diagrama.png)

- `ALUNOS (matricula PK, nome, email, turma)`
- `LIVROS (id_livro PK, titulo, autor, genero, ano_publicacao)`
- `EMPRESTIMOS (id_emprestimo PK, matricula FK → ALUNOS, id_livro FK → LIVROS, data_emprestimo, data_prevista, data_devolucao)`

O script de criação está em [`migrations/000001_create_tables.up.sql`](migrations/000001_create_tables.up.sql). As migrations são aplicadas automaticamente ao iniciar a aplicação.

## Executando no Linux

### Opção 1 — Docker (recomendado)

Requisitos: Docker e Docker Compose.

```bash
git clone https://github.com/RodrigoColi/biblioteca-escolar.git
cd biblioteca-escolar
docker compose up -d db
docker compose run --rm app
```

Se a porta 5432 já estiver em uso na máquina, escolha outra: `DB_PORT=55432 docker compose up -d db`.

Para apagar o banco e recomeçar do zero: `docker compose down -v`.

### Opção 2 — Go instalado localmente

Requisitos: Go 1.27+ e um PostgreSQL acessível.

```bash
# banco via Docker (ou use um PostgreSQL já instalado)
docker compose up -d db

export DATABASE_URL="postgres://biblioteca:biblioteca@localhost:5432/biblioteca?sslmode=disable"
go run ./cmd/biblioteca
```

Para gerar o binário: `go build -o biblioteca ./cmd/biblioteca && ./biblioteca`.

## Testes

| Tipo | Onde | Precisa de banco |
|------|------|------------------|
| Unitários (serviços com fakes em memória, `input`, `sqlutil`) | `internal/**/*_test.go` | não |
| Integração dos repositórios e relatórios | `internal/*/postgres_test.go` | sim |
| Ponta a ponta (binário real dirigido pela entrada padrão) | `e2e/` | sim |

```bash
# somente testes unitários (os que precisam de banco são ignorados)
go test ./...

# suíte completa: cada pacote usa um schema próprio, recriado a cada teste
docker compose up -d db
TEST_DATABASE_URL="postgres://biblioteca:biblioteca@localhost:5432/biblioteca?sslmode=disable" go test ./...
```

## Utilização

Ao iniciar, a aplicação exibe a splash screen com o total de registros de cada tabela e o autor do trabalho. Menu principal:

1. **Relatórios**
   - Empréstimos por gênero — total, em aberto e atrasados agrupados por gênero (`GROUP BY`)
   - Empréstimos detalhados — junção de empréstimos, alunos e livros (`JOIN`)
2. **Inserir registros** — escolha a entidade e informe cada atributo
3. **Remover registros** — escolha a entidade, selecione o registro pela chave e confirme. Se houver empréstimos vinculados, o sistema pergunta se deseja removê-los também
4. **Atualizar registros** — selecione o registro e informe os novos valores (Enter mantém o valor atual). O registro atualizado é exibido ao final
5. **Sair**

Após cada inserção, remoção ou atualização, o sistema pergunta se deseja realizar a operação novamente.
