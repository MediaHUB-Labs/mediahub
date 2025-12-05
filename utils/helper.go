package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Relative name of the Logs folder
const LogsDirName = "Logs"

func LogToFile(msg any) {
	// Create or truncate a File
	doLog := os.Getenv("LOGGER")
	// Dont log things to file if logging is disabled
	if doLog != "" && doLog != "true" {
		return
	}

	// --- FIX START: Determine the absolute path relative to the executable ---
	// Get the path to the currently running executable
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	// Get the directory containing the executable
	exeDir := filepath.Dir(exePath)

	// Build the full path for the Logs folder
	logBaseDir := filepath.Join(exeDir, LogsDirName)

	// Create the directories recursively
	if err := os.MkdirAll(logBaseDir, 0755); err != nil {
		fmt.Printf("Error creating Logs directory at %s: %v\n", logBaseDir, err)
		return
	}
	// --- FIX END ---

	// Try to open the File
	filenameTimeFormat := "2006-01-02"

	// Build the full absolute path to the log file
	logFilename := filepath.Join(logBaseDir, fmt.Sprintf("debug_%s.log", time.Now().Format(filenameTimeFormat)))

	file, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	// Log to File
	// Format the log line: [Timestamp] Message
	logLine := fmt.Sprintf("[%s] %v\n", time.Now().Format("2006-01-02 03:04:05 PM"), msg)

	// Log the formatted line to the File
	_, err = file.WriteString(logLine)
	if err != nil {
		fmt.Println("Error writing log message:", err)
		return
	}
}
