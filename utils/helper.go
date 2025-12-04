package utils

import (
	"fmt"
	"os"
	"time"
)

func LogToFile(msg any) {
	// Create or truncate a File
	doLog := os.Getenv("LOGGER")
	// Dont log things to file if logging is disabled
	if doLog != "true" {
		return
	}

	err := os.MkdirAll("Logs", 0755) // 0755 is the standard permission
	if err != nil {
		fmt.Println("Error creating Logs directory:", err)
		return
	}

	// Try to open the File
	filenameTimeFormat := "2006-01-02"
	logFilename := fmt.Sprintf("Logs/debug_%s.log", time.Now().Format(filenameTimeFormat))

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
