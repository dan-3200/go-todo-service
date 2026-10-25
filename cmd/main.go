package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"app/internal/application/usecase"
	"app/internal/infrastructure/config"
	httpadapter "app/internal/infrastructure/http"
	"app/internal/infrastructure/http/middleware"
	"app/internal/infrastructure/logging"
	"app/internal/infrastructure/postgres"
	"app/internal/infrastructure/telemetry"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	settings, err := config.Load()
	logger := logging.New(settings.Telemetry.ServiceName)
	slog.SetDefault(logger)
	if err != nil {
		logger.WarnContext(appCtx, "falha ao carregar arquivo .env; usando variáveis do ambiente",
			slog.Any("error", err),
		)
	}

	if settings.Telemetry.Enabled {
		cleanup, err := telemetry.SetupTelemetry(appCtx, settings.Telemetry)
		if err != nil {
			telemetry.LogTelemetryInitialization(appCtx, logger, settings.Telemetry.CollectorAddr, err)
		} else {
			logger.InfoContext(appCtx, "OpenTelemetry inicializado",
				slog.String("collector_addr", settings.Telemetry.CollectorAddr),
			)
			defer func() {
				if err := cleanup(); err != nil {
					logger.Error("falha ao desligar telemetria", slog.Any("error", err))
				}
			}()
		}
	}

	authMiddleware, err := middleware.NewJWT(settings.Auth.JWTSecret)
	if err != nil {
		logger.ErrorContext(appCtx, "falha ao inicializar autenticação", slog.Any("error", err))
		return
	}

	connection, err := postgres.Open(settings.Postgres)
	if err != nil {
		logger.ErrorContext(appCtx, "falha ao conectar com o banco de dados", slog.Any("error", err))
		return
	}
	defer connection.Close()

	repository, err := postgres.NewTodoRepository(connection)
	if err != nil {
		logger.ErrorContext(appCtx, "falha ao inicializar repositório", slog.Any("error", err))
		return
	}

	service := usecase.NewTodoService(repository)
	handler := httpadapter.NewTodoHandler(service)
	router := httpadapter.NewServer(handler, connection, authMiddleware)
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
			logger.ErrorContext(appCtx, "falha ao iniciar servidor HTTP", slog.Any("error", err))
		}
	case <-appCtx.Done():
		logger.InfoContext(appCtx, "desligando servidor HTTP")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(shutdownCtx, "falha ao desligar servidor HTTP", slog.Any("error", err))
		}
	}
}
