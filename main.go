package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	// Auth module
	"mediahub/auth/handler"
	"mediahub/auth/repository"
	"mediahub/auth/service"

	// Media module
	mediahandler "mediahub/media/handler"
	mediarepository "mediahub/media/repository"
	mediaservice "mediahub/media/service"

	// Upload module
	uploadrepository "mediahub/upload/repository"

	// Progress module
	progresshandler "mediahub/progress/handler"
	progressrepository "mediahub/progress/repository"
	progressservice "mediahub/progress/service"

	// Playlist module
	playlisthandler "mediahub/playlist/handler"
	playlistrepository "mediahub/playlist/repository"
	playlistservice "mediahub/playlist/service"

	// Transcode module
	"mediahub/transcode"

	// Shared
	"mediahub/dto"
	"mediahub/models"
	"mediahub/routes"
	"mediahub/utils"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using default values")
	}

	// ═══════════════════════════════════════════════════════
	// 1️ VALIDATE CONFIGURATION
	// ═══════════════════════════════════════════════════════

	// Read ENV values
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9123" // fallback if env not set
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "mediahub.db" // fallback if env not set
	}

	// Validate JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" || jwtSecret == "change-this-to-a-strong-secret-key" || jwtSecret == "secrete-key-here" {
		log.Println("⚠️  WARNING: JWT_SECRET is not set or using default value. Set a strong secret in .env for production!")
	}

	// Determine upload base path
	uploadBasePath := os.Getenv("UPLOAD_PATH")
	if uploadBasePath == "" {
		exePath, exeErr := os.Executable()
		if exeErr == nil {
			uploadBasePath = filepath.Join(filepath.Dir(exePath), "uploads")
		} else {
			uploadBasePath = "uploads"
		}
	}

	// ═══════════════════════════════════════════════════════
	// 2️ DATABASE SETUP
	// ═══════════════════════════════════════════════════════

	// Initialize SQLite DB
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate all models
	dbErr := db.AutoMigrate(
		&models.User{},
		&models.Media{},
		&models.UserMediaProgress{},
		&models.Playlist{},
		&models.PlaylistItem{},
	)
	if dbErr != nil {
		log.Fatalf("Fatal: Database migration failed: %v", dbErr)
	}
	fmt.Println("Database connected and migrated")

	// ═══════════════════════════════════════════════════════
	// 3️ CREATE INSTANCES (Dependency Injection)
	// ═══════════════════════════════════════════════════════

	// Auth
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Media
	mediaRepo := mediarepository.NewMediaRepository(db)
	uploadRepo := uploadrepository.NewUploadRepository(db)
	mediaService := mediaservice.NewMediaService(mediaRepo, uploadRepo)
	mediaHandler := mediahandler.NewMediaHandler(mediaService)

	// Progress
	progressRepo := progressrepository.NewProgressRepository(db)
	progressService := progressservice.NewProgressService(progressRepo, mediaRepo)
	progressHandler := progresshandler.NewProgressHandler(progressService)

	// Playlist
	playlistRepo := playlistrepository.NewPlaylistRepository(db)
	playlistService := playlistservice.NewPlaylistService(playlistRepo, mediaRepo)
	playlistHandler := playlisthandler.NewPlaylistHandler(playlistService)

	// Transcode
	transcodeService := transcode.NewTranscodeService(mediaRepo, uploadBasePath)

	// Log FFmpeg availability
	if utils.IsFFmpegAvailable() {
		fmt.Println("✅ FFmpeg detected — transcoding and thumbnails enabled")
	} else {
		fmt.Println("⚠️  FFmpeg not found — transcoding and thumbnails will be skipped")
	}

	// ═══════════════════════════════════════════════════════
	// 4️ GIN SETUP & ROUTES
	// ═══════════════════════════════════════════════════════

	// Set environment mode
	gin.SetMode(os.Getenv("GIN_MODE"))

	// Create router
	router := gin.Default()

	// Set max upload size (512 MB)
	router.MaxMultipartMemory = 512 << 20

	// CORS
	router.Use(CORSMiddleware())

	// Register all routes
	routes.RegisterAuthRoutes(router, userHandler)
	routes.RegisterMediaRoutes(router, mediaHandler, transcodeService)
	routes.RegisterProgressRoutes(router, progressHandler)
	routes.RegisterPlaylistRoutes(router, playlistHandler)

	// SERVER HEALTH CHECK
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.HealthReponse{
			Message: "OK",
		})
	})

	// ═══════════════════════════════════════════════════════
	// 5️ STATIC FILE SERVING
	// ═══════════════════════════════════════════════════════

	// Serve uploaded files (for direct access to images, thumbnails, etc.)
	router.Static("/uploads", uploadBasePath)

	// Serve transcoded HLS segments
	transcodedPath := filepath.Join(uploadBasePath, "transcoded")
	if err := os.MkdirAll(transcodedPath, os.ModePerm); err != nil {
		log.Printf("Warning: couldn't create transcoded directory: %v", err)
	}
	router.Static("/transcoded", transcodedPath)

	// --- SERVE UI ASSETS ---
	router.Static("/view", "./mediahub-ui/view")
	router.Static("/src", "./mediahub-ui/src")
	router.Static("/assets", "./mediahub-ui/assets")

	// Serve the root-level JS files from the UI folder
	router.StaticFile("/App.js", "./mediahub-ui/App.js")
	router.StaticFile("/output.css", "./mediahub-ui/assets/output.css")

	// --- FALLBACK ---
	router.NoRoute(func(c *gin.Context) {
		c.File("./mediahub-ui/index.html")
	})

	// ═══════════════════════════════════════════════════════
	// 6️ START SERVER
	// ═══════════════════════════════════════════════════════

	fmt.Printf("\n🚀 MediaHUB server starting on http://localhost:%s\n", port)
	fmt.Printf("📁 Upload path: %s\n", uploadBasePath)
	fmt.Printf("🔑 JWT secret: %s\n\n", func() string {
		if len(jwtSecret) > 4 {
			return jwtSecret[:4] + "****"
		}
		return "****"
	}())

	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Server failed:", err)
	}

}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Change * to your specific domain for better security
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
