# Go ToDo API - Clean Architecture
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
[![Postman](https://img.shields.io/badge/Postman-FF6C37?style=for-the-badge&logo=postman&logoColor=white)](https://www.postman.com/dan-3200/workspace/publico/collection/43029232-dfa83f1a-5ff2-47f7-a0ab-7cf1d0adb96c?action=share&creator=43029232)

## Descrição 

API RESTful para gerenciamento de tarefas (ToDos), desenvolvida em Golang com o framework HTTP Gin e banco de dados PostgreSQL, inspirada no padrão Clean Architecture. Permite criar, listar, atualizar e remover tarefas por meio de endpoints REST.

O projeto foi criado com foco em aprendizado, aplicando princípios de design como SOLID e o padrão Singleton.

## Tecnologias
- `Golang 1.27`
- `Gin (framework http)`
- `PostgreSQL (pgx)`
- `SQLc`
- `Atlas`
- `Docker`

## Integração contínua

O workflow GitHub Actions valida as migrations Atlas, confere se o schema desejado e
o código SQLc gerado estão sincronizados, executa testes, `go vet` e build da aplicação.
A imagem Docker é construída apenas em pushes na branch `main`.

## Estrutura do projeto (Clean Architecture)
```
├── internal
│   ├── domain                   # Regras de negócio
│   │   ├── entity               # Entidades
│   │   └── vo                   # Value objects
│   ├── application              # Orquestração da aplicação
│   │   ├── dto                  # Modelos de entrada e saída
│   │   ├── ports                # Contratos dos adapters
│   │   └── usecase              # Casos de uso
│   └── infrastructure           # Adapters externos
│       ├── config               # Configuração de ambiente
│       ├── http                 # Handlers, DTOs, rotas e servidor Gin
│       └── postgres             # Adapter PostgreSQL e código SQLc gerado
│
├── database                      # Fonte de verdade SQL
│   ├── schema.sql                # Schema PostgreSQL
│   ├── migrations                # Migrations Atlas versionadas e atlas.sum
│   └── queries/todo.sql          # Queries usadas pelo SQLc
│
├── atlas.hcl                     # Ambientes e diretório de migrations Atlas
├── sqlc.yaml                     # Configuração de geração SQLc
│
├── tests                         # Testes centralizados por camada
│   ├── application
│   ├── domain
│   └── infrastructure
│
├── docker                        # Configuração de containers
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── cmd/main.go                  # Ponto de entrada e composition root
├── go.mod                       # Dependências
├── go.sum                       # Hash das dependências
├── .env.example                 # Variáveis de ambiente
├── Makefile                     # Automação
└── readme.md
```
---

O diretório ``domain`` concentra entidades e value objects sem depender de banco, HTTP ou frameworks.
``application`` separa DTOs, portas e casos de uso, enquanto
``infrastructure`` implementa os adaptadores Gin e PostgreSQL. O adapter PostgreSQL
usa código tipado gerado pelo SQLc a partir dos arquivos em ``database``. O
`database/schema.sql` representa o estado desejado, enquanto o Atlas gera e aplica o
histórico executável de `database/migrations`.

O pool PostgreSQL é configurado por `POSTGRES_MAX_OPEN_CONNS`,
`POSTGRES_MAX_IDLE_CONNS`, `POSTGRES_CONN_MAX_LIFETIME` e
`POSTGRES_CONN_MAX_IDLE_TIME`. Os valores padrão são 10, 2, 30 minutos e
5 minutos, respectivamente.

Quando `OTEL_ENABLED=true`, a aplicação exporta traces, métricas de runtime e logs para o
collector OTLP/gRPC configurado em `OTEL_EXPORTER_OTLP_ENDPOINT`. A conexão, o tempo de
desligamento e a taxa de amostragem podem ser configurados por `OTEL_COLLECTOR_INSECURE`,
`OTEL_SHUTDOWN_TIMEOUT` e `OTEL_TRACER_RATIO`.

As dependências apontam para dentro: infraestrutura depende da aplicação e do domínio,
e a aplicação depende apenas do domínio.

### Entity Todo
```golang
type Todo struct {
	ID        vo.TodoID
	Title     string
	Content   string
	Status    bool
	CreatedAt time.Time
}
```

## Instrução de Instalação 
```bash
git clone https://github.com/DAN-3200/go-todo-service.git
cd go-todo-service
cp .env.example .env
make db-main
set -a && . ./.env && set +a
make migrate-apply
make run
```

> [!IMPORTANT]
> Instale o Atlas CLI v1.3.0 antes de usar os comandos de migration. A API não altera
> o schema durante o startup.

> [!IMPORTANT]
> Se o banco já contém o schema anterior à adoção do Atlas, não execute a baseline.
> Registre-a uma única vez com `make migrate-baseline VERSION=20261025110000`.

## Variáveis de ambiente

Use `.env.example` como base e não versione o `.env`. Durações usam o formato Go,
como `15s`, `5m` ou `1h`.

| Variável | Obrigatória | Padrão | Descrição |
| --- | --- | --- | --- |
| `API_PORT` | Não | `8200` | Porta HTTP da aplicação. |
| `API_READ_HEADER_TIMEOUT` | Não | `5s` | Limite para leitura dos headers. |
| `API_READ_TIMEOUT` | Não | `15s` | Limite para leitura da requisição. |
| `API_WRITE_TIMEOUT` | Não | `15s` | Limite para escrita da resposta. |
| `API_IDLE_TIMEOUT` | Não | `60s` | Limite para conexões ociosas. |
| `JWT_SECRET` | Sim | — | Segredo HS256 com no mínimo 32 bytes. |
| `POSTGRES_USER` | Sim | — | Usuário do PostgreSQL. |
| `POSTGRES_PASSWORD` | Sim | — | Senha do PostgreSQL. |
| `POSTGRES_DB` | Sim | — | Nome do banco de dados. |
| `POSTGRES_HOST` | Sim | — | Host usado pela aplicação. |
| `POSTGRES_PORT` | Sim | — | Porta usada pela aplicação. |
| `POSTGRES_MAX_OPEN_CONNS` | Não | `10` | Máximo de conexões abertas. |
| `POSTGRES_MAX_IDLE_CONNS` | Não | `2` | Máximo de conexões ociosas. |
| `POSTGRES_CONN_MAX_LIFETIME` | Não | `30m` | Vida máxima de uma conexão. |
| `POSTGRES_CONN_MAX_IDLE_TIME` | Não | `5m` | Tempo máximo ocioso da conexão. |
| `ATLAS_DATABASE_URL` | Sim para Atlas local | — | URL PostgreSQL usada por `migrate-apply`, `status` e `baseline`. |
| `OTEL_ENABLED` | Não | `false` | Habilita OpenTelemetry. |
| `OTEL_SERVICE_NAME` | Não | `go-todo-service` | Nome do serviço observado. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Não | `localhost:4317` | Endpoint gRPC do collector. |
| `OTEL_COLLECTOR_INSECURE` | Não | `true` | Usa conexão OTLP sem TLS. |
| `OTEL_SHUTDOWN_TIMEOUT` | Não | `5s` | Limite para desligar providers. |
| `OTEL_TRACER_RATIO` | Não | `1.0` | Taxa de amostragem entre zero e um. |

## Fluxo SQLc e Atlas

Instale a versão usada pelo CI:

```bash
curl -sSf https://atlasgo.sh | ATLAS_VERSION=v1.3.0 sh
atlas version
```

Para uma alteração de banco:

1. Edite `database/schema.sql` e, quando necessário, `database/queries`.
2. Execute `make generate` e revise o código SQLc.
3. Execute `make migrate-diff NAME=<nome>` e revise o SQL e `atlas.sum`.
4. Execute `make migrate-validate`.
5. Execute `make migrate-apply` no banco local.
6. Versione schema, queries, migration, `atlas.sum` e código SQLc gerado juntos.

Comandos disponíveis:

```bash
make generate                         # gerar código SQLc
make sqlc-vet                         # validar schema e queries SQLc
make migrate-diff NAME=add_due_date   # gerar migration Atlas
make migrate-validate                 # validar checksum e replay
make migrate-apply                    # aplicar migrations pendentes
make migrate-status                   # consultar versão aplicada
make migrate-baseline VERSION=20261025110000
make db-schema                        # alias para migrate-apply
```

`ATLAS_DATABASE_URL` precisa estar exportada no shell. Ao usar o `.env.example`
sem alterações de formato, carregue-o com:

```bash
set -a && . ./.env && set +a
```

> [!IMPORTANT]
> Para execução do docker, certifique-se de configurar o arquivo `.env` com base no `.env.example`.

> [!IMPORTANT]
> Os testes de integração PostgreSQL usam Testcontainers e iniciam uma instância
> efêmera de `postgres:16-alpine`. Portanto, `go test ./...` requer Docker disponível.


## API Endpoints
| Route | Descrição | Autenticação |
| --- | --- | --- |
| <kbd>POST /todo</kbd> | Criar uma nova tarefa | JWT |
| <kbd>GET /todo/:id</kbd> | Buscar tarefa específica por ID | JWT |
| <kbd>GET /todo</kbd> | Listar todas as tarefas | JWT |
| <kbd>PATCH /todo/:id</kbd> | Editar uma tarefa existente por ID | JWT |
| <kbd>DELETE /todo/:id</kbd> | Remover uma tarefa existente por ID | JWT |
| <kbd>GET /healthz</kbd> | Verificar a saúde da aplicação | Pública |
| <kbd>GET /readiness</kbd> | Verificar aplicação e PostgreSQL | Pública |

As rotas protegidas exigem `Authorization: Bearer <token>` com um JWT HS256 válido e
não expirado. A API apenas valida tokens emitidos externamente; ela não possui endpoint
de login ou emissão de tokens.

## Formato das respostas

Toda resposta da API usa um envelope com o status HTTP, timestamp e dados ou erro.

```json
{
  "success": true,
  "status": 200,
  "data": {},
  "timestamp": "2026-10-24T12:00:00Z"
}
```

```json
{
  "success": false,
  "status": 400,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "título é obrigatório"
  },
  "timestamp": "2026-10-24T12:00:00Z"
}
```

> [Link publico do Postman](https://www.postman.com/dan-3200/workspace/publico/collection/43029232-dfa83f1a-5ff2-47f7-a0ab-7cf1d0adb96c?action=share&creator=43029232)

## Licença
Este projeto está licenciado sob a Licença MIT. Consulte o arquivo [LICENSE](./LICENSE) para mais detalhes.
