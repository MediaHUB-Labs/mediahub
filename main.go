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

	// Register all routes
	routes.RegisterAuthRoutes(router, userHandler)
	routes.RegisterMediaRoutes(router, mediaHandler)

	// SERVER HEALTH CHECK
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.HealthReponse{
			Message: "OK",
		})
	})

	// Start Server
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Server failed:", err)
	}

}
