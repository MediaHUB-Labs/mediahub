package routes

import (
	"mediahub/auth/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, userHandler *handler.UserHandler) {

	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/login", userHandler.Login)

		auth.POST("/user", userHandler.GetUser)
		auth.PUT("/user", userHandler.UpdateUser)
		auth.DELETE("/user", userHandler.DeleteUser)
	}

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
