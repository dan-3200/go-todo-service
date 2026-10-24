package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type DatabasePinger interface {
	Ping() error
}

func NewServer(handler *TodoHandler, database DatabasePinger) *gin.Engine {
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	registerTodoRoutes(engine, handler)

	engine.GET("/readiness", func(ctx *gin.Context) {
		if database.Ping() != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"postgres": false})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":   "ready",
			"postgres": true,
		})
	})

	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return engine
}
