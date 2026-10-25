package http

import "github.com/gin-gonic/gin"

func registerTodoRoutes(engine *gin.Engine, handler *TodoHandler, authMiddleware gin.HandlerFunc) {
	todo := engine.Group("/todo")
	todo.Use(authMiddleware)
	{
		todo.POST("", handler.Create)
		todo.GET("/:id", handler.Get)
		todo.GET("", handler.List)
		todo.PATCH("/:id", handler.Update)
		todo.DELETE("/:id", handler.Delete)
	}
}
