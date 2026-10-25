package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"app/internal/infrastructure/config"
)

func Open(ctx context.Context, settings config.PostgresConfig) (*sql.DB, error) {
	uri := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		settings.User,
		settings.Password,
		settings.Host,
		settings.Port,
		settings.Database,
	)

	connection, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com o banco de dados: %w", err)
	}

	connection.SetMaxOpenConns(settings.MaxOpenConns)
	connection.SetMaxIdleConns(settings.MaxIdleConns)
	connection.SetConnMaxLifetime(settings.ConnMaxLifetime)
	connection.SetConnMaxIdleTime(settings.ConnMaxIdleTime)

	if err := connection.PingContext(ctx); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("erro ao pingar o banco de dados: %w", err)
	}

	return connection, nil
}
