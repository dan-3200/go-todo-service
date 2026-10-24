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
- `Docker`

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
│   └── queries/todo.sql          # Queries usadas pelo SQLc
│
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
└── README.md
```
---

O diretório ``domain`` concentra entidades e value objects sem depender de banco, HTTP ou frameworks.
``application`` separa DTOs, portas e casos de uso, enquanto
``infrastructure`` implementa os adaptadores Gin e PostgreSQL. O adapter PostgreSQL
usa código tipado gerado pelo SQLc a partir dos arquivos em ``database``.

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
make db-main
make db-schema
go run ./cmd
```

> [!IMPORTANT]
> A API não cria mais a tabela no startup. Aplique `database/schema.sql` com
> `make db-schema` antes de iniciar o serviço, inclusive ao usar um volume PostgreSQL existente.
## Instrução de uso

```bash
cd go-todo-service
make <tag>
```
---
```Makefile
# <tag>

# executar codigo
run: 
   go run ./cmd

# hot-reload
dev: 
	air

# rodar todos os tests (unitários e integração)
test: 
	go test ./...

# gerar código SQLc versionado
generate:
	go tool sqlc generate

# validar schema e queries com SQLc
sqlc-vet:
	go tool sqlc vet

# rodar banco principal no docker
db-main:
	docker compose -f 'docker/docker-compose.yml' up -d --build 'database'

 # rodar banco para tests (integração) no docker
db-test:
	docker compose -f 'docker/docker-compose.yml' up -d --build 'db-test'

# aplicar schema no banco principal
db-schema:
	docker compose -f 'docker/docker-compose.yml' exec -T database sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < database/schema.sql
```
> [!IMPORTANT]
> Para execução do docker, certifique-se de configurar o arquivo `.env` com base no `.env.example`.


## API Endpoints
| Route                   | Descrição                                  |
|-------------------------|---------------------------------------------|
| <kbd>POST /todo</kbd>     | Criar uma nova tarefa                     |
| <kbd>GET /todo/:id</kbd>  | Buscar tarefa específica por ID          |
| <kbd>GET /todo</kbd>      | Listar todas as tarefas                  |
| <kbd>PATCH /todo/:id</kbd>| Editar uma tarefa existente por ID   |
| <kbd>DELETE /todo/:id</kbd>| Remover uma tarefa existente por ID     |

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
