package database

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// DB holds the global database connection instance.
var DB *gorm.DB

// Card represents the MTG card model for GORM
type Card struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	OracleID   *string  `gorm:"type:varchar(255)"`
	Name       string   `gorm:"type:varchar(500);not null"`
	ManaCost   *string  `gorm:"type:varchar(100)"`
	CMC        *float64 `gorm:"type:decimal(10,2)"`
	TypeLine   string   `gorm:"type:varchar(500);not null"`
	OracleText *string  `gorm:"type:text"`
	Power      *string  `gorm:"type:varchar(20)"`
	Toughness  *string  `gorm:"type:varchar(20)"`
	Loyalty    *string  `gorm:"type:varchar(20)"`

	// Array fields - use pq.StringArray
	Colors        pq.StringArray `gorm:"type:text[]"`
	ColorIdentity pq.StringArray `gorm:"type:text[]"`
	Keywords      pq.StringArray `gorm:"type:text[]"`

	// JSON fields stored as text
	CardFaces  *string `gorm:"type:jsonb"`
	ImageURIs  *string `gorm:"type:jsonb"`
	Legalities *string `gorm:"type:jsonb"`
	Prices     *string `gorm:"type:jsonb"`

	SetCode         string  `gorm:"type:varchar(50);not null"`
	SetName         *string `gorm:"type:varchar(500)"`
	CollectorNumber *string `gorm:"type:varchar(50)"`
	Rarity          string  `gorm:"type:varchar(50);not null"`
	Artist          *string `gorm:"type:varchar(500)"`
	FlavorText      *string `gorm:"type:text"`
	ReleasedAt      *string `gorm:"type:varchar(50)"`
	Lang            string  `gorm:"type:varchar(10);default:''"`
	CachedAt        int64   `gorm:"type:bigint;default:0"`
}

// InitializeDatabase connects to the PostgreSQL database and performs migrations.
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

func GetDB() *gorm.DB {
	return DB
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

func GetRandomCard() (Card, error) {
	var card Card
	result := DB.Raw("SELECT * FROM cards TABLESAMPLE BERNOULLI(1) WHERE lang = 'en' LIMIT 1").Scan(&card)
	return card, result.Error
}

func SearchFuzzyOracleText(name string, text []string) ([]Card, error) {
	var out []Card
	var lastErr error

	for _, val := range text {
		var cards []Card

		DB.Exec("SELECT set_limit(0.3)")

		// Query with similarity matching
		result := DB.Not("name = ?", name).
			Distinct("name").
			Where("oracle_text % ?", val).
			Order(gorm.Expr("similarity(oracle_text, ?) DESC", val)). // Use gorm.Expr
			Limit(50).
			Find(&cards)
		// result := DB.Not("name = ?", name).Distinct("name").Where("oracle_text ILIKE ?", "%"+val+"%").Limit(50).Find(&cards)
		if result.Error != nil {
			lastErr = result.Error
			continue
		}
		if len(cards) > 0 {
			out = append(out, cards...)
		}
	}

	if len(out) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return out, nil
}

// GetCardByID retrieves a card by its Scryfall ID
func GetCardByID(id string) (*Card, error) {
	var card Card
	result := DB.Select("Name", "ImageURIs", "Colors", "CardFaces", "OracleText", "ManaCost").Where("id = ?", id).First(&card)
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

// PrimeDatabase streams a large JSON file and batch inserts cards
func PrimeDatabase(file io.Reader) error {
	decoder := json.NewDecoder(file)

	// Read opening bracket
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read opening bracket: %w", err)
	}

	const batchSize = 1000
	var cards []*Card
	var totalCount int
	var batchCount int

	startTime := time.Now()
	fmt.Println("Starting database priming...")

	// Read array elements
	for decoder.More() {
		var rawCard map[string]interface{}
		if err := decoder.Decode(&rawCard); err != nil {
			return fmt.Errorf("failed to decode card: %w", err)
		}

		card := mapScryfallToCard(rawCard)
		cards = append(cards, card)

		// When batch is full, insert
		if len(cards) >= batchSize {
			if err := batchInsertCards(cards); err != nil {
				return fmt.Errorf("failed to insert batch: %w", err)
			}

			totalCount += len(cards)
			batchCount++

			// Progress update every 10 batches (10,000 cards)
			if batchCount%10 == 0 {
				elapsed := time.Since(startTime)
				rate := float64(totalCount) / elapsed.Seconds()
				fmt.Printf("Inserted %d cards (%.0f cards/sec)...\n", totalCount, rate)
			}

			cards = cards[:0] // Reset slice
		}
	}

	// Insert remaining cards
	if len(cards) > 0 {
		if err := batchInsertCards(cards); err != nil {
			return fmt.Errorf("failed to insert final batch: %w", err)
		}
		totalCount += len(cards)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\n✓ Priming complete! Inserted %d cards in %s (%.0f cards/sec)\n",
		totalCount, elapsed.Round(time.Second), float64(totalCount)/elapsed.Seconds())

	return nil
}

// batchInsertCards inserts or updates a batch of cards using upsert
func batchInsertCards(cards []*Card) error {
	// Use Clauses with OnConflict to handle duplicates
	// This will update existing records instead of failing
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, // Conflict on primary key
		DoUpdates: clause.AssignmentColumns([]string{
			"updated_at", "oracle_id", "name", "mana_cost", "cmc",
			"type_line", "oracle_text", "power", "toughness", "loyalty",
			"colors", "color_identity", "keywords", "card_faces",
			"image_uris", "legalities", "prices", "set_code", "set_name",
			"collector_number", "rarity", "artist", "flavor_text",
			"released_at", "lang", "cached_at",
		}),
	}).CreateInBatches(cards, len(cards)).Error
}

// mapScryfallToCard converts Scryfall JSON to Card model
func mapScryfallToCard(data map[string]interface{}) *Card {
	card := &Card{}

	// Required fields
	if id, ok := data["id"].(string); ok {
		card.ID = id
	}
	if name, ok := data["name"].(string); ok {
		card.Name = name
	}
	if typeLine, ok := data["type_line"].(string); ok {
		card.TypeLine = typeLine
	}
	if rarity, ok := data["rarity"].(string); ok {
		card.Rarity = rarity
	}
	if setCode, ok := data["set"].(string); ok {
		card.SetCode = setCode
	}

	// Optional fields
	if oracleID, ok := data["oracle_id"].(string); ok {
		card.OracleID = &oracleID
	}
	if manaCost, ok := data["mana_cost"].(string); ok {
		card.ManaCost = &manaCost
	}
	if cmc, ok := data["cmc"].(float64); ok {
		card.CMC = &cmc
	}
	if oracleText, ok := data["oracle_text"].(string); ok {
		card.OracleText = &oracleText
	}
	if power, ok := data["power"].(string); ok {
		card.Power = &power
	}
	if toughness, ok := data["toughness"].(string); ok {
		card.Toughness = &toughness
	}
	if loyalty, ok := data["loyalty"].(string); ok {
		card.Loyalty = &loyalty
	}
	if setName, ok := data["set_name"].(string); ok {
		card.SetName = &setName
	}
	if collectorNumber, ok := data["collector_number"].(string); ok {
		card.CollectorNumber = &collectorNumber
	}
	if artist, ok := data["artist"].(string); ok {
		card.Artist = &artist
	}
	if flavorText, ok := data["flavor_text"].(string); ok {
		card.FlavorText = &flavorText
	}
	if releasedAt, ok := data["released_at"].(string); ok {
		card.ReleasedAt = &releasedAt
	}
	if lang, ok := data["lang"].(string); ok {
		card.Lang = lang
	}

	// Initialize arrays
	card.Colors = pq.StringArray{}
	card.ColorIdentity = pq.StringArray{}
	card.Keywords = pq.StringArray{}

	// Arrays
	if colors, ok := data["colors"].([]interface{}); ok && len(colors) > 0 {
		card.Colors = make(pq.StringArray, 0, len(colors))
		for _, c := range colors {
			if color, ok := c.(string); ok {
				card.Colors = append(card.Colors, color)
			}
		}
	}

	if colorIdentity, ok := data["color_identity"].([]interface{}); ok && len(colorIdentity) > 0 {
		card.ColorIdentity = make(pq.StringArray, 0, len(colorIdentity))
		for _, c := range colorIdentity {
			if color, ok := c.(string); ok {
				card.ColorIdentity = append(card.ColorIdentity, color)
			}
		}
	}

	if keywords, ok := data["keywords"].([]interface{}); ok && len(keywords) > 0 {
		card.Keywords = make(pq.StringArray, 0, len(keywords))
		for _, k := range keywords {
			if keyword, ok := k.(string); ok {
				card.Keywords = append(card.Keywords, keyword)
			}
		}
	}

	// JSON fields
	if imageURIs, ok := data["image_uris"].(map[string]interface{}); ok {
		if jsonBytes, err := json.Marshal(imageURIs); err == nil {
			jsonStr := string(jsonBytes)
			card.ImageURIs = &jsonStr
		}
	}

	if cardFaces, ok := data["card_faces"].([]interface{}); ok {
		if jsonBytes, err := json.Marshal(cardFaces); err == nil {
			jsonStr := string(jsonBytes)
			card.CardFaces = &jsonStr
		}
	}

	if legalities, ok := data["legalities"].(map[string]interface{}); ok {
		if jsonBytes, err := json.Marshal(legalities); err == nil {
			jsonStr := string(jsonBytes)
			card.Legalities = &jsonStr
		}
	}

	if prices, ok := data["prices"].(map[string]interface{}); ok {
		if jsonBytes, err := json.Marshal(prices); err == nil {
			jsonStr := string(jsonBytes)
			card.Prices = &jsonStr
		}
	}

	card.CachedAt = time.Now().Unix()

	return card
}
