package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Temporarily load environment variables from .env file for local development
	// Only for go run scripts/init_db.go, not for the main application
	if _, err := os.Stat(".env"); err == nil {
		log.Println("Loading environment variables from .env file")
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	} else {
		log.Println(".env file not found, relying on environment variables")
	}

	// Get database connection details from environment or use defaults
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "lite_collector")

	// Construct DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPassword, dbHost, dbPort)
	log.Printf("Connecting to MySQL at %s:%s with user %s", dbHost, dbPort, dbUser)

	// Connect to MySQL server
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}
	defer db.Close()

	// Create database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
	log.Printf("Database %s created or already exists", dbName)

	// Close connection and reopen to the specific database
	db.Close()

	dsnWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = sql.Open("mysql", dsnWithDB)
	if err != nil {
		log.Fatalf("Error connecting to database %s: %v", dbName, err)
	}
	defer db.Close()

	// Read and execute init.sql
	sqlFile, err := os.Open("init.sql")
	if err != nil {
		log.Fatalf("Error opening init.sql: %v", err)
	}
	defer sqlFile.Close()

	// Create scanner to read line by line
	scanner := bufio.NewScanner(sqlFile)
	var statement strings.Builder
	inString := false
	stringChar := byte(0)
	// inComment := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comment lines
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		// Simple string literal detection to avoid splitting inside strings
		for i := 0; i < len(line); i++ {
			switch line[i] {
			case '"', '\'', '`':
				if !inString {
					inString = true
					stringChar = line[i]
				} else if line[i] == stringChar {
					inString = false
				}
			}
		}

		// Add line to current statement
		statement.WriteString(line)
		statement.WriteString("\n")

		// Check if line ends with semicolon and we're not in a string
		if strings.HasSuffix(line, ";") && !inString {
			// Execute the statement
			sql := strings.TrimSpace(statement.String())
			if sql != "" {
				log.Printf("Executing SQL: %s...", truncateString(sql, 100))
				_, err := db.Exec(sql)
				if err != nil {
					log.Fatalf("Error executing SQL: %v\nSQL: %s", err, sql)
				}
			}
			// Reset statement builder
			statement.Reset()
			inString = false
		}
	}

	// Check for any remaining statement
	if statement.Len() > 0 {
		sql := strings.TrimSpace(statement.String())
		if sql != "" && !inString {
			log.Printf("Executing SQL: %s...", truncateString(sql, 100))
			_, err := db.Exec(sql)
			if err != nil {
				log.Fatalf("Error executing SQL: %v\nSQL: %s", err, sql)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading init.sql: %v", err)
	}

	log.Println("Database initialized successfully")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// truncateString returns a truncated version of the string for logging
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}
