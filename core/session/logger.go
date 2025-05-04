package session

import (
    "log"
		"fmt"
    "os"
    "path/filepath"
)

func init() {
    // Get the user's home directory
    userDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("[ERROR] Could not determine user home directory: %s", err.Error())
    }

    // Path to the .oblivion directory
    oblivionDir := filepath.Join(userDir, ".oblivion")

    // Ensure the .oblivion directory exists
    if _, err := os.Stat(oblivionDir); os.IsNotExist(err) {
        if err := os.MkdirAll(oblivionDir, 0755); err != nil {
            log.Fatalf("[ERROR] Could not create .oblivion directory: %s", err.Error())
        }
    }

    // Open or create the session log file in the .oblivion directory
    logFilePath := filepath.Join(oblivionDir, "session.log")
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err == nil {
        log.SetOutput(logFile)
        logSession("[INFO] Logger initialized.")
    } else {
        log.Printf("[ERROR] Could not open log file: %s", err.Error())
    }
}

func (s *Session) logError(err error, context string) {
    if err != nil {
        log.Printf("[ERROR] %s: %s", context, err.Error())
    }
}

func logInfo(message string) {
    logSession(fmt.Sprintf("[INFO] %s",message))
}

func logSession(message string) {
	log.Println(message)
}

func logCommand(cmd string) {
    log.Println("[CMD]", cmd)
}
