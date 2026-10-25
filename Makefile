# !windows powershell: 'choco install make' 

# Evitar conflito com nome de arquivos
.PHONY: dev test db-main db-test db-schema run generate sqlc-vet migrate-diff migrate-validate migrate-apply migrate-status migrate-baseline

# executar código
run: 
	go run ./cmd

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

# gerar uma migration a partir de database/schema.sql
migrate-diff:
	@test -n "$(NAME)" || (echo "uso: make migrate-diff NAME=<nome>" && exit 1)
	atlas migrate diff "$(NAME)" --env local

# validar checksum e reproducao das migrations
migrate-validate:
	atlas migrate validate --env ci

# aplicar migrations pendentes no banco configurado
migrate-apply:
	atlas migrate apply --env local

# exibir o estado das migrations no banco configurado
migrate-status:
	atlas migrate status --env local

# registrar a baseline em um banco que ja possui o schema
migrate-baseline:
	@test -n "$(VERSION)" || (echo "uso: make migrate-baseline VERSION=<timestamp>" && exit 1)
	atlas migrate apply --env local --baseline "$(VERSION)"

# rodar banco principal no docker
db-main:
	docker compose -f 'docker/docker-compose.yml' up -d --build 'database'

# rodar banco para tests (integração) no docker
db-test:
	docker compose -f 'docker/docker-compose.yml' up -d --build 'db-test'

# alias compativel para aplicar migrations no PostgreSQL
db-schema: migrate-apply
