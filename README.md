# Go API template

Boilerplate REST em Go (1.22+) com Gin, GORM, PostgreSQL, Redis/asynq, JWT e migrations. Organização por feature em `internal/`.

## Pré-requisitos

- Go 1.22 ou superior
- Docker e Docker Compose (para Postgres e Redis locais)
- [golang-migrate](https://github.com/golang-migrate/migrate) instalado no PATH (`migrate`) para os alvos `migrate-*` do Makefile
- Opcional: [golangci-lint](https://golangci-lint.run/) para `make lint`

## VS Code / Cursor

Instale a extensão oficial do Go:

- [golang.go](https://marketplace.visualstudio.com/items?itemName=golang.Go)

Na primeira abertura do projeto, a extensão vai pedir para instalar as ferramentas Go necessárias. Aceite todas. As principais são:

| Ferramenta      | Função                                                  |
| --------------- | ------------------------------------------------------- |
| `gopls`         | Language server (autocomplete, navegação, erros inline) |
| `goimports`     | Formatação + organização de imports no save             |
| `golangci-lint` | Linter agregado (rode também via `make lint`)           |
| `dlv`           | Debugger                                                |

O arquivo `.vscode/settings.json` já está configurado para formatar e lint ao salvar.

## Primeira execução

Na raiz do repositório:

```bash
cd /Users/caiohenrique/go-api-template && go mod tidy && cp .env.example .env && docker compose up -d && make migrate-up && make run
```

Em outro terminal, com o mesmo `.env` e a partir da mesma raiz:

```bash
cd /Users/caiohenrique/go-api-template && make run-worker
```

O comando `go mod tidy` gera o `go.sum` e baixa dependências (necessário antes do primeiro `docker build` ou `make build`).

## Variáveis de ambiente

Copie [`.env.example`](.env.example) para `.env`. O Makefile carrega `.env` para `migrate-up` / `migrate-down` (via `include`).

## Migrations

Criar nova migration:

```bash
make migrate-create name=add_orders_table
```

Aplicar / reverter:

```bash
make migrate-up
make migrate-down
```

## API e worker

- **API HTTP**: `make run` (porta `APP_PORT`, padrão `8080`)
- **Worker asynq**: `make run-worker` (consome jobs no Redis)

Build de binários:

```bash
make build
# gera bin/api e bin/worker
```

## Rotas

| Método | Rota          | Auth       | Descrição                              |
| ------ | ------------- | ---------- | -------------------------------------- |
| POST   | `/users`      | não        | Cria usuário                           |
| GET    | `/users/me`   | Bearer JWT | Retorna o usuário autenticado          |
| POST   | `/auth/login` | não        | Login (email + senha) → `access_token` |

## Exemplos com curl

Criar usuário:

```bash
curl -sS -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"longpassword"}'
```

Login:

```bash
curl -sS -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"longpassword"}'
```

Perfil (substitua o token):

```bash
curl -sS http://localhost:8080/users/me \
  -H "Authorization: Bearer <access_token>"
```

## Docker (imagem da aplicação)

O [`Dockerfile`](Dockerfile) gera os binários `api` e `worker`. A imagem inicia a API por padrão. Para rodar o worker com a mesma imagem:

```bash
docker run --rm --env-file .env --entrypoint /app/worker <sua-imagem>
```

## Testes e lint

```bash
make test
make lint
```

## Detalhe de arquitetura

O ID do usuário autenticado é colocado no `context.Context` da requisição em [`internal/platform/requestctx`](internal/platform/requestctx) (evita ciclo de importação entre `auth` e `user`). O middleware em `internal/features/auth` grava esse valor após validar o JWT.
