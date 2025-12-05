package routes

import (
	"mediahub/auth/handler"
	mediaHandler "mediahub/media/handler"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.Engine, userHandler *handler.UserHandler) {

	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", userHandler.Signup)
		auth.POST("/login", userHandler.Login)

		auth.POST("/user", userHandler.GetUser)
		auth.PUT("/user", userHandler.UpdateUser)
		auth.DELETE("/user", userHandler.DeleteUser)
	}

}

func RegisterMediaRoutes(router *gin.Engine, mediaHandler *mediaHandler.MediaHandler) {

	// MEDIA ROUTES
	media := router.Group("/api/media")
	{
		media.GET("/health", mediaHandler.MediaHealth)

		// // Creation
		// media.POST("/add", mediaHandler.CreateMedia)

		// // Fetching/Listing (Uses query parameters like ?category=Movie or ?limit=20)
		// media.GET("/list", mediaHandler.ListMedia)
		// media.GET("/search", mediaHandler.SearchMedia)
		// media.GET("/categories", mediaHandler.GetCategories)

		// // Get Details (ID is in the request body)
		// media.POST("/details", mediaHandler.GetMediaDetails)
		// // Update Metadata (ID is in the request body)
		// media.PUT("/metadata", mediaHandler.UpdateMediaMetadata)
		// // Deletion (ID is in the request body)
		// media.DELETE("/item", mediaHandler.DeleteMedia)
		// // Serving (ID is in the request body)
		// media.POST("/play", mediaHandler.StreamMedia)
	}

}

// TODO------------------------------------------------
// USER STATE ROUTES
// progress := router.Group("/api/progress")
// {
// 	// Save/Update Progress (User ID and Media ID are in the request body)
// 	progress.POST("/save", progressHandler.SaveProgress)
// 	// List Progress (User ID is obtained from the JWT token/context)
// 	progress.GET("/continue", progressHandler.ListContinueWatching)
// 	// Clear Progress (Media ID is in the request body)
// 	progress.DELETE("/clear", progressHandler.RemoveProgress)
// }
