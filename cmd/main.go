package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"app/internal/application/usecase"
	"app/internal/infrastructure/config"
	httpadapter "app/internal/infrastructure/http"
	"app/internal/infrastructure/postgres"
	"app/internal/infrastructure/telemetry"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	settings, err := config.Load()
	if err != nil {
		log.Println("[.env error]: ", err)
	}

	if settings.Telemetry.Enabled {
		cleanup, err := telemetry.SetupTelemetry(appCtx, settings.Telemetry)
		if err != nil {
			log.Println("[error ao inicializar telemetria]:", err)
		} else {
			defer func() {
				if err := cleanup(); err != nil {
					log.Println("[error ao desligar telemetria]:", err)
				}
			}()
		}
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
	router := httpadapter.NewServer(handler, connection)
	server := &http.Server{
		Addr:              ":" + settings.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("[error ao iniciar servidor]:", err)
		}
	case <-appCtx.Done():
		log.Println("[desligando servidor HTTP]")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Println("[erro ao desligar servidor HTTP]:", err)
		}
	}
}
