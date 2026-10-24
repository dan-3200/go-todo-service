package main

import (
	"log"

	"app/internal/application/usecase"
	"app/internal/infrastructure/config"
	httpadapter "app/internal/infrastructure/http"
	"app/internal/infrastructure/postgres"
)

func main() {
	settings, err := config.Load()
	if err != nil {
		log.Println("[.env error]: ", err)
	}

	connection, err := postgres.Open(settings.Postgres)
	if err != nil {
		log.Println("[error ao conectar com o banco de dados]:", err)
		return
	}
	defer connection.Close()

	repository, err := postgres.NewTodoRepository(connection)
	if err != nil {
		log.Println("[error ao inicializar repositório]:", err)
		return
	}

	service := usecase.NewTodoService(repository)
	handler := httpadapter.NewTodoHandler(service)
	server := httpadapter.NewServer(handler, connection)

	if err := server.Run(":" + settings.Server.Port); err != nil {
		log.Println("[error ao iniciar servidor]:", err)
	}
}
