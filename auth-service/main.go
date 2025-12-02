package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// --- CONFIGURATION ---
type Config struct {
	AuthServicePort int
	MediahubDBPath  string
}

var AppConfig Config

func loadEnv() {
	// If your .env is at infrastructure/.env
	err := godotenv.Load("../infrastructure/.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func initConfig() {
	AppConfig.MediahubDBPath = os.Getenv("MEDIAHUB_DB_PATH")
	if AppConfig.MediahubDBPath == "" {
		log.Fatalf("Fatal: MEDIAHUB_DB_PATH environment variable not set.")
	}

	portStr := os.Getenv("AUTH_SERVICE_PORT")
	if portStr == "" {
		log.Fatalf("Fatal: AUTH_SERVICE_PORT environment variable not set.")
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Fatal: AUTH_SERVICE_PORT must be an integer, got: %s", portStr)
	}
	AppConfig.AuthServicePort = portInt
}

// ----------------------

// Global database handle
var db *sql.DB

// --- DATABASE INITIALIZATION AND SCHEMA SETUP ---
func initDB() {
	var err error
	db, err = sql.Open("sqlite", AppConfig.MediahubDBPath)

	if err != nil {
		log.Fatalf("Error opening database file at %s: %v", AppConfig.MediahubDBPath, err)
	}

	if pingErr := db.Ping(); pingErr != nil {
		log.Fatalf("Database connection check (Ping) failed: %v", pingErr)
	}

	fmt.Println("Database connection established successfully.")

	// Create the Users Table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	fmt.Println("Users table initialized successfully.")
}

// ------------------------------------

// Define structs for request/response
type StatusMessage struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Password Hashing Function
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Handler function for the POST /signup endpoint.
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqData SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		if err == io.EOF {
			http.Error(w, "Request body cannot be empty.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Invalid request format.", http.StatusBadRequest)
		return
	}

	// 1. Hash the password securely
	hashedPassword, err := hashPassword(reqData.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error during processing.", http.StatusInternalServerError)
		return
	}

	// 2. Prepare and execute the SQL INSERT statement
	insertSQL := `INSERT INTO users (email, password_hash) VALUES (?, ?)`

	_, err = db.Exec(insertSQL, reqData.Email, hashedPassword)

	if err != nil {
		// Simple check for UNIQUE constraint violation
		if err.Error() == "UNIQUE constraint failed: users.email" {
			http.Error(w, "User with this email already exists.", http.StatusConflict) // HTTP 409
			return
		}
		log.Printf("Error inserting user: %v", err)
		http.Error(w, "Internal server error during registration.", http.StatusInternalServerError)
		return
	}

	// 3. Send success response (HTTP 201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	successResponse := StatusMessage{Status: "success", Service: "User successfully created."}
	if err := json.NewEncoder(w).Encode(successResponse); err != nil {
		log.Printf("Failed to encode success response: %v", err)
	}
}

func main() {
	// 1. Initialize dependencies
	loadEnv()
	initConfig()
	initDB()
	defer db.Close()

	port := AppConfig.AuthServicePort

	// 2. Register handlers
	listenAddr := fmt.Sprintf(":%d", port)
	http.HandleFunc("/signup", signupHandler)

	fmt.Printf("Auth Service starting on http://localhost%s...\n", listenAddr)

	// 3. Start the server
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
