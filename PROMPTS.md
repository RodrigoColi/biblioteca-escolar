# Uso de IA

Registro dos prompts utilizados no desenvolvimento, conforme exigido pelo edital.

Os prompts abaixo foram reescritos em português a partir das sessões com o Claude Code, consolidando as instruções dadas em cada etapa do desenvolvimento.

## 1. Planejamento e arquitetura do projeto

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:** o plano e a arquitetura foram revisados e aprovados antes de qualquer código. A partir deles, foram definidos o tema (Biblioteca Escolar), as tabelas `ALUNOS`, `LIVROS` e `EMPRESTIMOS`, os dois relatórios, o fluxo dos menus, a organização em pacotes por funcionalidade com camadas e o processo de migrations.

**Prompt:**

> Leia o edital do trabalho C2 de Banco de Dados (PDF anexo) e o projeto de exemplo do professor (`example_crud_oracle`) e monte um plano para desenvolver a aplicação em Go com PostgreSQL. Antes de planejar, sugira alguns temas do edital com as tabelas e os relatórios de cada um para eu escolher.
>
> **Escopo do plano**
> - Modelo relacional: tabelas, atributos, tipos, PK, FK, obrigatoriedade e cardinalidades.
> - Dois relatórios: um com `GROUP BY` e outro com `JOIN`.
> - Fluxo dos menus de inserir, remover e atualizar conforme o item 6 do edital.
> - Splash screen com a contagem de registros de cada tabela (`select count(1) total_x from x`).
> - Entregáveis do item 7 e riscos de prazo (entrega em 30/10/2026).
>
> **Regras do edital que a arquitetura deve respeitar**
> - Sem ORM nem query builder: o SQL é montado por concatenação.
> - A remoção de um registro referenciado por FK é tratada pela aplicação (avisar ou perguntar se remove os filhos), sem `ON DELETE CASCADE`.
> - Entidades associadas guardam a instância da outra entidade (um empréstimo contém o `Aluno` e o `Livro`, não só as chaves).
> - Execução no Linux.
>
> **Arquitetura: camadas dentro de pacotes por funcionalidade (estilo DDD)**
>
> Cada entidade tem o seu pacote (`aluno`, `livro`, `emprestimo`, além de `relatorio`), e dentro dele as camadas ficam em arquivos separados:
> 1. **Domínio (`<entidade>.go`):** a struct da entidade, o método `String()` e os erros do domínio (`ErrNotFound`, erros de validação). Não depende de nenhuma outra camada.
> 2. **Repositório (`repository.go`):** apenas a interface com as operações que o serviço precisa do banco: contar, verificar existência, listar, buscar por chave, inserir, atualizar, remover e, para alunos e livros, contar e remover os empréstimos vinculados.
> 3. **Persistência (`postgres.go`):** a implementação da interface para PostgreSQL com `database/sql` e o driver `pgx`. O SQL é concatenado usando funções de um pacote `sqlutil` (escape de aspas simples, formatação de datas, `NULL` para datas vazias). `sql.ErrNoRows` vira `ErrNotFound` do domínio. A remoção em cascata feita pela aplicação roda em uma transação. Garanta em tempo de compilação que a struct implementa a interface.
> 4. **Serviço (`service.go`):** as regras de negócio, sem SQL. Verifica a existência antes de inserir, atualizar ou remover, valida os campos e as datas, impede emprestar um livro que já está em aberto e verifica dependências antes de remover. O serviço de empréstimos depende de interfaces pequenas, definidas no próprio pacote (`AlunoFinder`, `LivroFinder`), para buscar aluno e livro sem acoplar aos repositórios deles.
> 5. **Interface com o usuário (`internal/cli`):** splash screen, menus, telas e o laço "deseja realizar mais alguma operação?". Conversa só com os serviços e usa um pacote `internal/input` para ler e validar a entrada (reapresentar a pergunta em caso de valor inválido, Enter para manter o valor atual, confirmação S/N).
>
> As dependências seguem uma única direção: `cli` → serviço → interface do repositório ← implementação PostgreSQL, sem ciclos de importação. A montagem fica em `cmd/biblioteca/main.go`, nesta ordem: lê a configuração (`DATABASE_URL`, com valor padrão), aplica as migrations, abre a conexão, cria os repositórios e os serviços e inicia a CLI.
>
> **Migrations e scripts SQL**
> - Use golang-migrate com arquivos numerados em `migrations/`: `000001_create_tables` (tabelas, PK, FK e restrições `CHECK` das datas) e `000002_seed` (carga inicial para demonstração), cada um com o seu `.up.sql` e `.down.sql`.
> - O `000001_create_tables.up.sql` também é o entregável "script de criação das tabelas e relacionamentos" do edital, então deve ser legível e executável sozinho.
> - Os arquivos `.sql` são embutidos no binário com `embed` e aplicados automaticamente ao iniciar a aplicação (`Up`, ignorando quando não há mudança). Assim, reiniciar não reaplica a carga inicial, e o `.down.sql` desfaz cada etapa.
> - Livros e empréstimos usam chave gerada pelo banco (`IDENTITY`). A matrícula do aluno é informada pelo usuário.

## 2. Containerização com Docker

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:** o `Dockerfile`, o `.dockerignore` e o `docker-compose.yml` gerados foram adicionados ao repositório. A aplicação foi executada com `docker compose up -d db` e `docker compose run --rm app`, e a suíte de testes foi executada contra o banco do container.

**Prompt:**

> Crie a containerização do projeto Biblioteca Escolar (Go 1.27, PostgreSQL, migrations embutidas no binário), para que o professor rode tudo em uma máquina Linux apenas com Docker, sem instalar Go nem PostgreSQL.
>
> **Dockerfile (multi-stage)**
> 1. **Estágio de build:** `golang:1.27-alpine`. Copie primeiro o `go.mod` e o `go.sum` e rode `go mod download`, para que o cache de dependências só seja refeito quando elas mudarem. Depois copie o código e compile `./cmd/biblioteca` com `CGO_ENABLED=0` e `-ldflags="-s -w"`, gerando um binário estático e menor.
> 2. **Estágio final:** `alpine:3.22` com apenas o binário. Instale `tzdata` e defina `TZ=America/Sao_Paulo`, porque a situação dos empréstimos (em aberto ou atrasado) depende da data atual. Use `ENTRYPOINT` com o binário.
>
> **`.dockerignore`:** exclua o que não é necessário para compilar (`.git`, `docs`, arquivos Markdown e o binário local).
>
> **`docker-compose.yml`**
> - Serviço `db`: `postgres:17-alpine` com usuário, senha e banco `biblioteca` e o mesmo fuso horário. Os dados ficam em um volume nomeado. A porta do host é configurável por `DB_PORT` (padrão 5432), caso a 5432 já esteja em uso. O healthcheck usa `pg_isready`.
> - Serviço `app`: compilado a partir do Dockerfile, com `DATABASE_URL` apontando para o host `db`. Deve depender do `db` com `condition: service_healthy`, para que as migrations só rodem com o banco pronto, e usar `stdin_open` e `tty`, já que a aplicação é um menu interativo.
>
> O uso esperado é `docker compose up -d db` seguido de `docker compose run --rm app`. `docker compose down -v` apaga o banco para recomeçar com a carga inicial.

## 3. Suíte de testes (unitários, integração e ponta a ponta)

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:** os testes gerados foram adicionados ao repositório (`*_test.go`, `internal/testdb/`, `e2e/`). Para conferir se os testes detectam erros de verdade, foram introduzidas falhas propositais no código (testes de mutação), e os testes acusaram cada uma delas. Durante a geração, `internal/input` foi ajustado para permitir substituir a função de saída (`os.Exit`) nos testes. A suíte completa (89 testes) foi executada contra o banco do `docker-compose.yml`, sem falhas e sem testes ignorados.

**Prompt:**

> Escreva a suíte de testes automatizados do projeto Biblioteca Escolar (Go, PostgreSQL, golang-migrate), respeitando a arquitetura existente: pacotes de domínio `aluno`, `livro`, `emprestimo` e `relatorio` (cada um com `repository.go` definindo a interface, `postgres.go` com a implementação e `service.go` com as regras), além de `sqlutil`, `input`, `cli`, `database` e `cmd/biblioteca`.
>
> Restrições gerais:
> - Use apenas o pacote `testing` da biblioteca padrão — sem frameworks de asserção, sem mocks gerados, sem testcontainers. Dublês de teste devem ser fakes em memória escritos à mão que implementam as interfaces `Repository`, `AlunoFinder` e `LivroFinder`.
> - Prefira testes table-driven quando houver várias entradas para a mesma regra. Nomes de testes descritivos, poucos comentários.
> - Os testes não podem depender da data atual: quando a situação de um empréstimo importar (em aberto/atrasado), gere as datas relativas a hoje com folga de vários dias.
> - `go test ./...` sem banco configurado deve passar, pulando (`t.Skip`) os testes que precisam do PostgreSQL. Com `TEST_DATABASE_URL` definida, todos devem rodar. Cada pacote deve usar seu próprio schema do PostgreSQL (recriado e migrado a cada teste), para que pacotes rodem em paralelo sem interferência.
> - Depois de escrever os testes, verifique se eles realmente detectam erros: introduza falhas propositais nas regras principais (validações, verificação de dependências, cálculo de atraso) e confirme que algum teste falha em cada caso.
>
> 1. **Unitários (sem banco)**
>    - `sqlutil`: escape de aspas simples (incluindo tentativas de SQL injection como `'); DROP TABLE alunos;--`), formatação de datas, `NULL` para data nula, conversão de `sql.NullTime`.
>    - `input`: reprompt em campo obrigatório vazio, inteiro inválido, opção fora do intervalo, data inválida (inclusive 31/02); Enter mantém o valor atual; `-` limpa data opcional; confirmação aceita S/SIM/N/NÃO em qualquer caixa e reprompta o resto; fim da entrada (EOF) chama a função de saída.
>    - `emprestimo.Situacao`: devolvido, atrasado, vence hoje (não atrasado), em aberto; `String()` com e sem data de devolução.
>    - Serviços de `aluno` e `livro`: cada regra de validação; criação duplicada; atualização e remoção de registro inexistente; remoção com empréstimos vinculados sem cascata (erro e nada removido) e com cascata (remove tudo); propagação de erros do repositório.
>    - Serviço de `emprestimo`: aluno ou livro inexistente; data prevista ou de devolução anterior ao empréstimo; datas iguais permitidas; livro já emprestado em aberto; empréstimo já devolvido não bloqueia o livro; atualizar o próprio empréstimo em aberto não conflita consigo mesmo; trocar para um livro emprestado é bloqueado; aluno e livro completos são carregados no empréstimo.
> 2. **Integração dos repositórios (`postgres.go`, banco real)**
>    - CRUD completo de cada entidade, incluindo IDs gerados por identity, `ErrNotFound` ao buscar inexistente, textos com aspas gravados literalmente, datas nulas e não nulas.
>    - Contagens usadas pela splash screen, `CountEmprestimos` e `CountOpenByLivro` (com o ID ignorado).
>    - `DeleteWithEmprestimos` remove filhos e pai e não afeta outros registros; remover pai com filhos sem cascata falha pela FK; constraints de data rejeitam valores inválidos.
>    - Relatórios: agrupamento por gênero (incluindo gênero sem empréstimos com zeros, em aberto e atrasados corretos) e junção detalhada com a situação calculada.
>    - `database.Migrate` é idempotente e a carga inicial gera 5 alunos, 6 livros e 6 empréstimos.
> 3. **Ponta a ponta (binário real)**
>    - Compile `cmd/biblioteca`, execute-o como subprocesso com `DATABASE_URL` apontando para um schema de teste e alimente a entrada padrão com roteiros completos. Verifique a saída e o estado final do banco.
>    - Cubra todos os itens do edital: splash screen com contagens e integrantes; os dois relatórios; inserir em cada entidade com loop "inserir mais algum?" (Sim volta ao menu de entidades, Não volta ao principal); remover com confirmação, cancelamento voltando ao menu principal, registro com FK recusando e aceitando a remoção dos filhos; atualizar mantendo valores com Enter, exibindo o registro atualizado, preenchendo e limpando a data de devolução.
>    - Casos de borda: entradas inválidas reapresentam a pergunta, opção inexistente no menu, matrícula duplicada, registro inexistente, nome com aspas e tentativa de SQL injection, fim inesperado da entrada (sai com código 0), banco indisponível (sai com código diferente de 0 e mensagem clara) e reinício da aplicação sem reaplicar a carga inicial.
>    - Cada execução deve ter timeout para que um loop de reprompt nunca trave a suíte.
>
> Atualize o README com os comandos para rodar os testes com e sem banco.

## 4. README do projeto

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:** o README gerado foi revisado e adicionado ao repositório. Os comandos foram conferidos executando a aplicação pelos dois caminhos descritos: via Docker Compose e com Go local apontando para um PostgreSQL próprio.

**Prompt:**

> Escreva o `README.md` do projeto Biblioteca Escolar, em português, para o trabalho C2 de Banco de Dados (Prof. Howard Roatti, 2026/2, autor: Rodrigo Leite Colli Pedroso). O projeto é uma aplicação de console em Go com PostgreSQL, migrations via golang-migrate (aplicadas automaticamente ao iniciar) e Docker Compose.
>
> O README é avaliado pelo edital como "documentação explicativa de utilização do sistema no Linux", então deve permitir que alguém que nunca viu o projeto o execute em uma máquina Linux limpa seguindo apenas o texto. Seja direto e objetivo, sem texto de marketing.
>
> Seções:
> 1. **Apresentação:** uma frase sobre o sistema (alunos, livros e empréstimos, com relatórios) e a linha de autoria (autor, disciplina, professor, semestre). Inclua o link do vídeo demonstrativo no YouTube e use a URL real do repositório no comando `git clone`.
> 2. **Estrutura:** explique que o código é organizado por funcionalidade. Mostre em uma tabela o papel de cada arquivo dos pacotes de domínio (`<entidade>.go`, `repository.go`, `postgres.go`, `service.go`), deixando claro que o SQL é montado por concatenação, sem ORM, e depois a árvore de diretórios com uma descrição curta de cada pasta.
> 3. **Modelo relacional:** imagem `docs/diagrama.png`, as três tabelas em notação textual com PK e FK, e o link para o script de criação em `migrations/000001_create_tables.up.sql`.
> 4. **Executando no Linux:**
>    - Opção 1, Docker (recomendada): requisitos, `git clone`, `docker compose up -d db` e `docker compose run --rm app`. Explique como trocar a porta com `DB_PORT` se a 5432 estiver ocupada e como apagar o banco com `docker compose down -v`.
>    - Opção 2, Go instalado localmente: requisitos (versão do Go do `go.mod` e um PostgreSQL acessível), a variável `DATABASE_URL` com o valor padrão, `go run ./cmd/biblioteca` e como gerar o binário.
> 5. **Testes:** tabela com os tipos de teste (unitários, integração, ponta a ponta), onde ficam e se precisam de banco. Mostre os comandos com e sem `TEST_DATABASE_URL`.
> 6. **Utilização:** o que a splash screen mostra e o menu principal item a item (relatórios com GROUP BY e JOIN, inserir, remover com a pergunta sobre empréstimos vinculados, atualizar mantendo valores com Enter e exibindo o registro atualizado, sair), e o comportamento de repetir a operação ao final de cada ação.
>
> Use blocos de código para todos os comandos e confira cada comando com os arquivos reais do projeto (`docker-compose.yml`, `go.mod`, `internal/config`).

## 5. Diagrama relacional (Mermaid e PNG)

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:** o diagrama foi conferido coluna a coluna com o script `migrations/000001_create_tables.up.sql`. A versão em Mermaid foi salva em `docs/diagrama.mmd`, e a versão em imagem, gerada a partir do mesmo modelo, em `docs/diagrama.png`, que é a imagem referenciada no README.

**Prompt:**

> Gere o diagrama relacional do projeto Biblioteca Escolar em Mermaid (`erDiagram`), a partir do script de criação das tabelas em `migrations/000001_create_tables.up.sql` (PostgreSQL). O diagrama é um entregável avaliado no edital ("diagrama relacional, corretamente desenvolvido"), então precisa refletir exatamente o banco.
>
> - Tabelas `ALUNOS`, `LIVROS` e `EMPRESTIMOS`, com todos os campos, na ordem do script, cada um com o tipo SQL real (`INTEGER`, `VARCHAR(n)`, `DATE`).
> - Marque as chaves primárias com `PK` e as estrangeiras com `FK`, e indique em comentário de cada campo se é `NOT NULL` ou `NULL`, além de `IDENTITY` nas chaves geradas automaticamente.
> - Relacionamentos com a cardinalidade correta em notação pé de galinha: um aluno realiza zero ou muitos empréstimos, e um livro é emprestado em zero ou muitos empréstimos. Cada empréstimo pertence a exatamente um aluno e um livro. Use rótulos em português nos relacionamentos.
> - Não invente campos, tabelas ou restrições que não existam no script.
>
> Depois, gere a versão em imagem (PNG) do mesmo modelo para o repositório, legível em tamanho normal, incluindo uma legenda com o significado de PK, FK e da notação de cardinalidade, as chaves estrangeiras com as tabelas referenciadas (sem `ON DELETE CASCADE`) e as restrições `CHECK` das datas.

## 6. Ajustes finais e verificação da entrega

- **Fonte:** Claude Code (modelo Claude Opus 5.5), 28/10/2026
- **O que foi feito com a resposta:**
  - A splash screen e o README foram ajustados para o trabalho individual, com Rodrigo Leite Colli Pedroso como único integrante.
  - Os pontos pendentes apontados na revisão foram resolvidos: a imagem do diagrama, o README e o registro dos prompts.
  - A aplicação foi executada em Linux pelos dois caminhos do README: Docker Compose e Go local com PostgreSQL. A splash screen, os menus e os dois relatórios funcionaram com a carga inicial.
  - A suíte completa (89 testes) passou contra o banco do container.

**Prompt:**

> O trabalho será feito individualmente, apenas por Rodrigo Leite Colli Pedroso. Ajuste a splash screen e o README para refletir isso, mantendo o nome do professor, a disciplina e o semestre.
>
> Em seguida, revise o projeto contra o edital como se fosse o professor corrigindo:
> - Confira cada item da distribuição de pontos (relatórios, inserir, remover, atualizar, diagrama, interface, vídeo, README) e cada entregável do item 7. Para cada um, diga se está atendido, com a evidência no código ou no repositório, ou o que falta.
> - Confira as observações importantes: commits, `PROMPTS.md`, professor como colaborador, execução na VM da disciplina, prazo.
> - Execute a aplicação em Linux seguindo apenas o README, pelo caminho com Docker Compose e pelo caminho com Go local, e confirme que as migrations, a carga inicial, a splash screen e os relatórios funcionam.
> - Rode a suíte completa de testes contra o banco do Docker e informe o resultado.
>
> Termine com uma lista ordenada do que ainda falta fazer até a entrega em 30/10/2026, separando o que pode ser feito no código do que depende de mim.

