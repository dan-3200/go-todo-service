package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func OpenFromEnvironment() (*sql.DB, error) {
	uri := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	connection, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, fmt.Errorf("Erro ao conectar com o banco de dados: %v", err)
	}

	if err := connection.Ping(); err != nil {
		return nil, fmt.Errorf("Erro ao pingar o banco de dados: %v", err)
	}

	return connection, nil
}
