# !windows powershell: 'choco install make' 

# Evitar conflito com nome de arquivos
.PHONY: dev test db-main db-test db-schema run generate sqlc-vet

# executar código
run: 
	go run .

# hot-reload
dev: 
	air

# rodar todos os testes (unitários e integração)
test: 
	go test ./...

# gerar código de acesso ao banco a partir dos arquivos SQL
generate:
	go tool sqlc generate

# validar schema e queries SQL
sqlc-vet:
	go tool sqlc vet

# rodar banco principal no docker
db-main:
	docker compose -f 'docker-compose.yml' up -d --build 'database'

# rodar banco para tests (integração) no docker
db-test:
	docker compose -f 'docker-compose.yml' up -d --build 'db-test'

# aplicar schema SQL no PostgreSQL em execução
db-schema:
	docker compose exec -T database sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"' < database/schema.sql
