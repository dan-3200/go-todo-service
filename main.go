package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"app/internal/application/usecase"
	httpadapter "app/internal/infrastructure/http"
	"app/internal/infrastructure/postgres"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("[.env error]: ", err)
		return
	}
}

func main() {
	connection, err := postgres.OpenFromEnvironment()
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

	if err := repository.CreateSchema(context.Background()); err != nil {
		log.Println("[error ao criar tabela]:", err)
		return
	}

	service := usecase.NewTodoService(repository)
	handler := httpadapter.NewTodoHandler(service)
	server := httpadapter.NewServer(handler, connection)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8200"
	}

	if err := server.Run(":" + port); err != nil {
		log.Println("[error ao iniciar servidor]:", err)
	}
}
