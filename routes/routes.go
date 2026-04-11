package routes

import (
	"mediahub/auth"
	authHandler "mediahub/auth/handler"
	mediaHandler "mediahub/media/handler"
	playlistHandler "mediahub/playlist/handler"
	progressHandler "mediahub/progress/handler"
	"mediahub/transcode"
	"mediahub/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes sets up authentication routes.
// Login and Signup are public; user CRUD is protected by JWT middleware.
func RegisterAuthRoutes(router *gin.Engine, userHandler *authHandler.UserHandler) {

	authGroup := router.Group("/api/auth")
	{
		// Public routes
		authGroup.POST("/signup", userHandler.Signup)
		authGroup.POST("/login", userHandler.Login)

		// Protected routes (require JWT)
		protected := authGroup.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			protected.POST("/user", userHandler.GetUser)
			protected.PUT("/user", userHandler.UpdateUser)
			protected.DELETE("/user", userHandler.DeleteUser)
		}
	}
}

// RegisterMediaRoutes sets up media routes.
// Health check is public; all CRUD/streaming routes are protected by JWT middleware.
func RegisterMediaRoutes(router *gin.Engine, mediaHandler *mediaHandler.MediaHandler, transcodeService *transcode.TranscodeService) {

	media := router.Group("/api/media")
	{
		// Public routes
		media.GET("/health", mediaHandler.MediaHealth)

		// Protected routes (require JWT)
		protected := media.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			// Upload
			protected.POST("/add", mediaHandler.UploadMedia)

			// Listing & Search
			protected.GET("/list", mediaHandler.ListMedia)
			protected.GET("/search", mediaHandler.SearchMedia)
			protected.GET("/categories", mediaHandler.GetCategories)

			// Details & CRUD
			protected.POST("/details", mediaHandler.GetMediaDetails)
			protected.PUT("/metadata", mediaHandler.UpdateMediaMetadata)
			protected.DELETE("/item", mediaHandler.DeleteMedia)

			// Personal vault (photos, documents)
			protected.GET("/vault", mediaHandler.GetUserVault)

			// Streaming
			protected.GET("/stream/:id", mediaHandler.StreamMedia)

			// Thumbnails
			protected.GET("/thumbnail/:id", mediaHandler.ServeThumbnail)

			// HLS manifest
			protected.GET("/hls/:id", mediaHandler.ServeHLSManifest)

			// Transcoding
			if transcodeService != nil {
				protected.POST("/transcode/:id", func(c *gin.Context) {
					idStr := c.Param("id")
					id, err := strconv.ParseUint(idStr, 10, 64)
					if err != nil {
						c.JSON(http.StatusBadRequest, dto.ApiResponse{
							Success: false,
							Error:   "invalid media id",
						})
						return
					}

					if err := transcodeService.TranscodeToHLS(uint(id)); err != nil {
						c.JSON(http.StatusBadRequest, dto.ApiResponse{
							Success: false,
							Error:   err.Error(),
						})
						return
					}

					c.JSON(http.StatusAccepted, dto.ApiResponse{
						Success: true,
						Message: "Transcoding started in background",
					})
				})

				protected.GET("/transcode/status", func(c *gin.Context) {
					c.JSON(http.StatusOK, dto.ApiResponse{
						Success: true,
						Message: "Transcode status",
						Data:    transcodeService.GetStatus(),
					})
				})
			}
		}
	}
}

// RegisterProgressRoutes sets up user progress tracking routes (all protected).
func RegisterProgressRoutes(router *gin.Engine, progressHandler *progressHandler.ProgressHandler) {
	progress := router.Group("/api/progress")
	progress.Use(auth.AuthMiddleware())
	{
		progress.POST("/save", progressHandler.SaveProgress)
		progress.GET("/continue", progressHandler.ListContinueWatching)
		progress.DELETE("/clear", progressHandler.RemoveProgress)
	}
}

// RegisterPlaylistRoutes sets up playlist management routes (all protected).
func RegisterPlaylistRoutes(router *gin.Engine, playlistHandler *playlistHandler.PlaylistHandler) {
	playlist := router.Group("/api/playlist")
	playlist.Use(auth.AuthMiddleware())
	{
		playlist.POST("/create", playlistHandler.CreatePlaylist)
		playlist.GET("/list", playlistHandler.ListPlaylists)
		playlist.GET("/:id", playlistHandler.GetPlaylist)
		playlist.PUT("/:id", playlistHandler.UpdatePlaylist)
		playlist.DELETE("/:id", playlistHandler.DeletePlaylist)
		playlist.POST("/:id/add", playlistHandler.AddItem)
		playlist.DELETE("/:id/remove", playlistHandler.RemoveItem)
	}
}
