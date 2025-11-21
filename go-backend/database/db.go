package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB holds the global database connection instance.
var DB *gorm.DB

// InitializeDatabase connects to the database and performs migrations.
// It accepts a list of models to migrate.
func InitializeDatabase(models ...interface{}) {
	var err error
	// Use SQLite for simplicity in this base project.
	// Replace with postgres.Open() or mysql.Open() for production.
	DB, err = gorm.Open(sqlite.Open("inventory.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Database: Successfully connected.")

	// AutoMigrate: Creates or updates tables based on the provided models.
	err = DB.AutoMigrate(models...)
	if err != nil {
		log.Fatalf("Database: Failed to auto-migrate schema: %v", err)
	}
	fmt.Println("Database: Migration successful.")
}
