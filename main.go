package main

import (
	"fmt"
	"log"
	"mediahub/auth/handler"
	"mediahub/auth/repository"
	"mediahub/auth/service"
	"mediahub/dto"
	mediahandler "mediahub/media/handler"
	mediarepository "mediahub/media/repository"
	mediaservice "mediahub/media/service"
	"mediahub/models"
	"mediahub/routes"
	uploadrepository "mediahub/upload/repository"
	"net/http"
	"os"

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

	// Read ENV values
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "9123" // fallback if env not set
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "mediahub.db" // fallback if env not set
	}

	// Initialize SQLite DB
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate
	dbErr := db.AutoMigrate(&models.User{}, &models.Media{}, &models.UserMediaProgress{})
	if dbErr != nil {
		log.Fatalf("Fatal: Database migration failed: %v", dbErr)
	}
	fmt.Println("Database connected and migrated")

	// ═══════════════════════════════════════════════════════
	// 2️ CREATE INSTANCES (Direct, no interfaces)
	// ═══════════════════════════════════════════════════════

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	mediaRepo := mediarepository.NewMediaRepository(db)
	uploadRepo := uploadrepository.NewUploadRepository(db)
	mediaService := mediaservice.NewMediaService(mediaRepo, uploadRepo)
	mediaHandler := mediahandler.NewMediaHandler(mediaService)
	// ═══════════════════════════════════════════════════════
	// 3️ GIN SETUP & ROUTES
	// ═══════════════════════════════════════════════════════
	// Set environment mode
	gin.SetMode(os.Getenv("GIN_MODE"))
	// Routes
	router := gin.Default()

	// CORS
	router.Use(CORSMiddleware())

	// Register all routes
	routes.RegisterAuthRoutes(router, userHandler)
	routes.RegisterMediaRoutes(router, mediaHandler)

	// SERVER HEALTH CHECK
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.HealthReponse{
			Message: "OK",
		})
	})

	// --- 2. SERVE UI ASSETS ---
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

	// Start Server
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
