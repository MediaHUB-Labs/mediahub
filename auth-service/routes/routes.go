package routes

import (
	"mediahub/auth-service/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, userHandler *handler.UserHandler) {

	auth := router.Group("/auth")
	{
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/login", userHandler.Login)

		auth.GET("/users/:id", userHandler.GetUser)
		auth.PUT("/users/:id", userHandler.UpdateUser)
		auth.DELETE("/users/:id", userHandler.DeleteUser)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
