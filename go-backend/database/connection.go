package database

import (
	"context"
	"fmt"
	"go-backend/config"
	"log"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var GraphDriver neo4j.DriverWithContext

func InitializeDatabase(models ...interface{}) {
	cfg := config.PGConfig{}
    
    // 1. Parse env variables into struct
    if err := env.Parse(&cfg); err != nil {
        log.Fatalf("Database: Failed to parse config: %v", err)
    }

    // 2. The DSN String Template
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%d sslmode=%s search_path=public",
        cfg.Host,
        cfg.User,
        cfg.Pass,
        cfg.Name,
        cfg.Port,
        cfg.Ssl,
    )
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
	cfg := config.MGConfig{}
    if err := env.Parse(&cfg); err != nil {
        log.Fatalf("Memgraph: Failed to parse config: %v", err)
    }

    // Construct the URI
    uri := fmt.Sprintf("bolt://%s:%d", cfg.Host, cfg.Port)

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(cfg.User, cfg.Pass, ""))
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



func InitSystem(models ...interface{}) {
    // 1. Initialize Connections (These stay blocking as they are required)
    InitializeDatabase(models...)
    InitializeMemgraph()

    // 2. Perform Parity Check
    log.Println("Checking Database Parity")
    pgCount := GetPostgresCardCount()
    mgCount := GetMemgraphCardCount()

    fmt.Printf("Counts -> Postgres (Unique Cards): %d | Memgraph (Nodes): %d\n", pgCount, mgCount)

    if pgCount != mgCount || mgCount == 0 {
        fmt.Println("Out of sync! Background Memgraph re-sync started...")
        
        // START GOROUTINE: This prevents blocking the rest of the app
        go func() {
            if err := ReSyncToMemgraph(); err != nil {
                // Use log.Printf instead of Fatalf here so the app doesn't crash 
                // if only the graph sync fails
                log.Printf("Background re-sync failed: %v", err)
            } else {
                fmt.Println("Background Memgraph re-sync completed successfully!")
            }
        }()
        
    } else {
        fmt.Println("Databases are in sync. Ready to go!")
    }
}