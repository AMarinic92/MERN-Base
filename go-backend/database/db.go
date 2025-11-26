package database

import (
    "fmt"
    "log"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// DB holds the global database connection instance.
var DB *gorm.DB

// Card represents the MTG card model for GORM
type Card struct {
	gorm.Model
    ID              string         `gorm:"type:uuid;primaryKey" json:"id"`
    OracleID        *string        `gorm:"type:uuid" json:"oracle_id"`
    Name            string         `gorm:"type:varchar(255);not null;index" json:"name"`
    ManaCost        *string        `gorm:"type:varchar(50)" json:"mana_cost"`
    CMC             *float64       `gorm:"type:decimal(4,1)" json:"cmc"`
    TypeLine        string         `gorm:"type:varchar(255);not null;index" json:"type_line"`
    OracleText      *string        `gorm:"type:text" json:"oracle_text"`
    Power           *string        `gorm:"type:varchar(10)" json:"power"`
    Toughness       *string        `gorm:"type:varchar(10)" json:"toughness"`
    Loyalty         *string        `gorm:"type:varchar(10)" json:"loyalty"`
    Colors          []string       `gorm:"type:char(1)[];index:,type:gin" json:"colors"`
    ColorIdentity   []string       `gorm:"type:char(1)[]" json:"color_identity"`
    Keywords        []string       `gorm:"type:text[]" json:"keywords"`
    CardFaces       *string        `gorm:"type:jsonb" json:"card_faces"` // Store as JSON string
    SetCode         string         `gorm:"type:varchar(10);not null;index" json:"set_code"`
    SetName         *string        `gorm:"type:varchar(100)" json:"set_name"`
    CollectorNumber *string        `gorm:"type:varchar(20)" json:"collector_number"`
    Rarity          string         `gorm:"type:varchar(20);not null;index" json:"rarity"`
    ImageURIs       *string        `gorm:"type:jsonb" json:"image_uris"` // Store as JSON string
    Legalities      *string        `gorm:"type:jsonb" json:"legalities"` // Store as JSON string
    Prices          *string        `gorm:"type:jsonb" json:"prices"` // Store as JSON string
    Artist          *string        `gorm:"type:varchar(100)" json:"artist"`
    FlavorText      *string        `gorm:"type:text" json:"flavor_text"`
    ReleasedAt      *string        `gorm:"type:date" json:"released_at"`
    Lang            string         `gorm:"type:varchar(5);default:'en'" json:"lang"`
    CachedAt        int64          `gorm:"autoCreateTime" json:"cached_at"`
    UpdatedAt       int64          `gorm:"autoUpdateTime" json:"updated_at"`
}

// InitializeDatabase connects to the PostgreSQL database and performs migrations.
func InitializeDatabase(models ...interface{}) {
    var err error
    
    // Get PostgreSQL connection string from environment or use default
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // Default connection string - update with your credentials
        dsn = "host=localhost user=appuser password=@s$Fuck1337! dbname=postgres port=5432 sslmode=disable"
    }
    
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    fmt.Println("Database: Successfully connected to PostgreSQL.")
    
    // AutoMigrate: Creates or updates tables based on the provided models.
    err = DB.AutoMigrate(models...)
    if err != nil {
        log.Fatalf("Database: Failed to auto-migrate schema: %v", err)
    }
    fmt.Println("Database: Migration successful.")
}

// SearchCardByName searches for a card by exact name
func SearchCardByName(name string) (*Card, error) {
    var card Card
    result := DB.Where("name = ?", name).First(&card)
    if result.Error != nil {
        return nil, result.Error
    }
    return &card, nil
}

// SearchCardByNameFuzzy searches for cards with similar names (requires pg_trgm extension)
func SearchCardByNameFuzzy(name string) ([]Card, error) {
    var cards []Card
    result := DB.Where("name ILIKE ?", "%"+name+"%").Limit(10).Find(&cards)
    if result.Error != nil {
        return nil, result.Error
    }
    return cards, nil
}

// GetCardByID retrieves a card by its Scryfall ID
func GetCardByID(id string) (*Card, error) {
    var card Card
    result := DB.Where("id = ?", id).First(&card)
    if result.Error != nil {
        return nil, result.Error
    }
    return &card, nil
}

// UpsertCard inserts or updates a card (useful for caching Scryfall data)
func UpsertCard(card *Card) error {
    result := DB.Save(card)
    return result.Error
}