package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"app/internal/infrastructure/http/response"
)

type DatabasePinger interface {
	Ping() error
}

func NewServer(handler *TodoHandler, database DatabasePinger, authMiddleware gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	engine.Use(gin.Logger(), gin.CustomRecovery(handlePanic))
	engine.Use(suppressHealthCheckTelemetry())

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	registerTodoRoutes(engine, handler, authMiddleware)

	engine.GET("/readiness", func(ctx *gin.Context) {
		if database.Ping() != nil {
			response.Error(ctx, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "banco de dados indisponível", nil)
			return
		}

		response.Success(ctx, http.StatusOK, readinessResponse{Status: "ready", Postgres: true})
	})

	engine.GET("/healthz", func(ctx *gin.Context) {
		response.Success(ctx, http.StatusOK, healthResponse{Status: "healthy"})
	})

	engine.NoRoute(func(ctx *gin.Context) {
		response.Error(ctx, http.StatusNotFound, "ROUTE_NOT_FOUND", "rota não encontrada", nil)
	})

	engine.NoMethod(func(ctx *gin.Context) {
		response.Error(ctx, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "método não permitido", nil)
	})

	return engine
}

type healthResponse struct {
	Status string `json:"status"`
}

type readinessResponse struct {
	Status   string `json:"status"`
	Postgres bool   `json:"postgres"`
}
