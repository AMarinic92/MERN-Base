package database

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var GraphDriver neo4j.DriverWithContext

func InitializeDatabase(models ...interface{}) {
	dsn := "host=localhost user=appuser password=@s$Fuck1337! dbname=postgres port=5432 sslmode=disable search_path=public"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// Disable foreign key constraints during migration
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Database: Failed to connect: %v", err)
	}

	fmt.Println("Database: Connected successfully")

	// AutoMigrate should be safe - suppress the error if table exists
	if err := DB.AutoMigrate(models...); err != nil {
		// Check if it's just a "table already exists" error
		if !strings.Contains(err.Error(), "already exists") {
			log.Fatalf("Database: Failed to auto-migrate schema: %v", err)
		}
		fmt.Println("Database: Table already exists, continuing...")
	} else {
		fmt.Println("Database: Schema migrated successfully")
	}
}

func InitializeMemgraph() {
	uri := "bolt://localhost:7687" // Default Memgraph port
	user := ""                     // Default is empty for local
	password := ""

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		log.Fatalf("Memgraph: Failed to create driver: %v", err)
	}
	GraphDriver = driver

	// Ensure our "Lean Schema" Constraints/Indexes
	ctx := context.Background()
	executeSchema(ctx)
	fmt.Println("Memgraph: Connected and Schema Verified")
}

func GetDB() *gorm.DB {
	return DB
}