package main

import (
	"fmt"
	"log"
	"mediahub/auth/handler"
	"mediahub/auth/models"
	"mediahub/auth/repository"
	"mediahub/auth/routes"
	"mediahub/auth/service"
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
		dbPath = "mediahub.db"
	}

	// Initialize SQLite DB
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected:", dbPath)

	// Auto-migrate
	db.AutoMigrate(&models.User{})
	fmt.Println("✅ Database connected and migrated")

	// ═══════════════════════════════════════════════════════
	// 2️ CREATE INSTANCES (Direct, no interfaces)
	// ═══════════════════════════════════════════════════════

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// ═══════════════════════════════════════════════════════
	// 3️ GIN SETUP & ROUTES
	// ═══════════════════════════════════════════════════════

	router := gin.Default()

	// Register all routes
	routes.RegisterRoutes(router, userHandler)

	// Start Server
	log.Println("MediaHUB Server running on port:", port)
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal("Server failed:", err)
	}

}
