package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"lite-collector/models"
)

// DB is the global database connection
var DB *gorm.DB

// Init initializes the database connection
func Init(dataSourceName string) {
	var err error
	DB, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Form{},
		&models.Submission{},
		&models.SubmissionValue{},
		&models.BaseData{},
		&models.AIJob{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}