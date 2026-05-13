# Go API template

Boilerplate REST em Go (1.23+) com Gin, GORM, PostgreSQL, Redis/asynq, JWT e migrations. O código de negócio segue **Ports & Adapters (hexagonal)** com **package-by-feature**: cada feature em `internal/features/<nome>/` contém `domain/` (entidades puras), `app/` (casos de uso, ports e tipos `Input`/`View`) e `adapters/` (HTTP Gin, GORM, JWT, etc.). O wiring de implementações concretas fica em `cmd/api` e `cmd/worker`.

## Pré-requisitos

- Go 1.23 ou superior
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

## Arquitetura

Cada feature (`user`, `auth`, `order`) está em `internal/features/<feature>/` com três camadas em subpastas:

- **`domain/`** — entidades e erros de domínio, sem frameworks.
- **`app/`** — casos de uso, interfaces (ports) e tipos de entrada/saída do use case. Convenção de nomes: sufixo **`Input`** para entrada (ex.: `CreateUserInput`, `LoginInput`) e **`View`** para leitura (ex.: `UserView`, `OrderView`). Não usamos o termo “DTO” para esses tipos: eles modelam o caso de uso; quando o JSON HTTP é idêntico, as mesmas structs carregam tags `json:` / `validate:` / `form:` para binding direto nos handlers.
- **`adapters/`** — implementações: `adapters/http` (Gin, Swagger), `adapters/repo` (GORM), `adapters/jwt`, etc.

Features não importam o `app` ou `adapters` de outras features; comunicação cruzada usa ports (ex.: `auth/app` declara `UserFinder`, satisfeito pelo repositório de `user`). `auth/app` pode importar `user/domain` apenas para tipar retornos do port.

Utilitários compartilhados ficam em [`internal/platform`](internal/platform) (config, logger, `apperr`, paginação, fila asynq, bcrypt genérico, etc.).

O ID do usuário autenticado é colocado no `context.Context` da requisição em [`internal/platform/requestctx`](internal/platform/requestctx). O middleware em `internal/features/auth/adapters/http` grava esse valor após validar o JWT.
