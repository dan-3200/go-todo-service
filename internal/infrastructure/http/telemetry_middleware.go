package http

import (
	"github.com/gin-gonic/gin"

	"app/internal/infrastructure/logging"
)

func suppressHealthCheckTelemetry() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isHealthCheckPath(ctx.Request.URL.Path) {
			ctx.Request = ctx.Request.WithContext(logging.WithoutTelemetry(ctx.Request.Context()))
		}
		ctx.Next()
	}
}

func isHealthCheckPath(path string) bool {
	return path == "/healthz" || path == "/readiness"
}
