package database

import (
	"encoding/json"
	"fmt"
	"go-backend/models"
	"io"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)


func SearchCardByName(name string) (*models.Card, error) {
	var card models.Card
	result := DB.Where("name = ?", name).First(&card)
	if result.Error != nil {
		return nil, result.Error
	}
	return &card, nil
}


func batchInsertCards(cards []*models.Card) error {
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

func GetPostgresCardCount() int64 {
	var count int64
	// We count DISTINCT oracle_id because Memgraph uses OracleID as the unique anchor
	DB.Model(&models.Card{}).Distinct("oracle_id").Count(&count)
	return count
}

func ReSyncToMemgraph() error {
	// 1. Get the count of UNIQUE oracle_ids for an accurate progress bar
	var uniqueCount int64
	DB.Model(&models.Card{}).Distinct("oracle_id").Count(&uniqueCount)

	const batchSize = 1000
	// We use a raw query or GORM's Clauses to handle the DISTINCT ON requirement efficiently
	for i := 0; i < int(uniqueCount); i += batchSize {
		var cards []*models.Card
		
		// DISTINCT ON (oracle_id) ensures we only get one row per functional card.
		// We order by oracle_id (required by DISTINCT ON) and then released_at DESC 
		// to ensure the graph gets the most recent text/wording.
		err := DB.Raw(`
			SELECT DISTINCT ON (oracle_id) * FROM cards 
			WHERE oracle_id IS NOT NULL
			ORDER BY oracle_id, released_at DESC 
			LIMIT ? OFFSET ?`, batchSize, i).Scan(&cards).Error

		if err != nil {
			return fmt.Errorf("failed to fetch distinct cards: %w", err)
		}

		if err := syncToMemgraph(cards); err != nil {
			return fmt.Errorf("failed to sync batch to Memgraph: %w", err)
		}
		
		log.Printf("Synced %d/%d unique cards to Memgraph...", i+len(cards), uniqueCount)
	}

	log.Println(" Memgraph re-sync complete (Unique functional cards only).")
	return nil
}

// SearchCardByNameFuzzy searches for cards with similar names (requires pg_trgm extension)
func SearchCardByNameFuzzy(name string) ([]models.Card, error) {
    var cards []models.Card
    
    result := DB.Raw(`
        SELECT c.name, c.id, c.oracle_id, c.image_uris, c.colors, c.card_faces, c.oracle_text, c.mana_cost, c.cmc, c.color_identity, c.type_line
        FROM cards c
        INNER JOIN (
            SELECT name, MAX(id) as id
            FROM cards
            WHERE name % ? AND lang = 'en' AND deleted_at IS NULL 
            GROUP BY name
        ) as unique_cards ON c.name = unique_cards.name AND c.id = unique_cards.id
        ORDER BY similarity(c.name, ?) DESC
        LIMIT 10
    `, name, name).Scan(&cards)
    
    if result.Error != nil {
        return nil, result.Error
    }
    return cards, nil
}

func GetRandomCard() (models.Card, error) {
	var card models.Card
	result := DB.Raw("SELECT * FROM cards TABLESAMPLE BERNOULLI(1) WHERE lang = 'en' LIMIT 1").Scan(&card)
	return card, result.Error
}

func SearchFuzzyOracleText(name string, text []string) ([]models.Card, error) {
    var (
        mu      sync.Mutex
        wg      sync.WaitGroup
        out     []models.Card
        lastErr error
    )
    for _, val := range text {
        wg.Add(1)
        go func(searchVal string) {
            defer wg.Done()
            var cards []models.Card
            // Each goroutine gets its own DB session
            db := DB.Session(&gorm.Session{})
            db.Exec("SELECT set_limit(0.65)")
            
            result := db.Raw(`
                SELECT c.name, c.type_line, c.id, c.image_uris, c.colors, c.card_faces, c.oracle_text, c.color_identity, c.mana_cost, c.cmc
                FROM cards c
                INNER JOIN (
                    SELECT name, MAX(id) as id
                    FROM cards
                    WHERE name != ? 
                        AND lang = 'en' 
                        AND oracle_text % ? 
                        AND deleted_at IS NULL
						AND NOT type_line ILIKE '%Token%'
						AND NOT type_line ILIKE '%Emblem%'
						AND NOT type_line ILIKE 'Basic Land%'
                    GROUP BY name
                ) as unique_cards ON c.name = unique_cards.name AND c.id = unique_cards.id
                ORDER BY similarity(c.oracle_text, ?) DESC
                LIMIT 50
            `, name, searchVal, searchVal).Scan(&cards)
            
            mu.Lock()
            defer mu.Unlock()
            if result.Error != nil {
                lastErr = result.Error
                return
            }
            if len(cards) > 0 {
                out = append(out, cards...)
            }
        }(val)
    }
    wg.Wait()
    if len(out) == 0 && lastErr != nil {
        return nil, lastErr
    }
    return out, nil
}

// GetCardByID retrieves a card by its Scryfall ID
func GetCardByID(id string) (*models.Card, error) {
	var card models.Card
	result := DB.Select("ID","Name","TypeLine", "cmc","Power","Toughness", "ImageURIs", "Colors", "CardFaces", "OracleText","OracleID", "ManaCost", "ColorIdentity").Where("id = ?", id).First(&card)
	if result.Error != nil {
		return nil, result.Error
	}
	return &card, nil
}

func GetCardVariants(oracleID string, currentID string) ([]models.Card, error) {
    var variants []models.Card
    
    result := DB.Select(
        "id", "name", "type_line", "cmc", "power", "toughness", 
        "image_uris", "colors", "card_faces", "oracle_text", 
        "oracle_id", "mana_cost", "color_identity",
    ).Where("oracle_id = ?", oracleID).
      Where("id != ?", currentID).
      Where("lang = ?", "en").
      Find(&variants)

    if result.Error != nil {
        return nil, result.Error
    }

    return variants, nil
}
// UpsertCard inserts or updates a card (useful for caching Scryfall data)
func UpsertCard(card *models.Card) error {
	result := DB.Save(card)
	return result.Error
}


// PrimeDatabase streams a large JSON file and batch inserts cards
func PrimeDatabase(file io.Reader) error {
	decoder := json.NewDecoder(file)

	// Read opening bracket
	if _, err := decoder.Token(); err != nil {
		log.Fatalf("failed to read opening bracket: %w", err)
		return fmt.Errorf("failed to read opening bracket: %w", err)
	}

	const batchSize = 1000
	var cards []*models.Card
	var totalCount int
	var batchCount int

	startTime := time.Now()
	log.Println("Starting database priming...")

	// Read array elements
	for decoder.More() {
		var rawCard map[string]interface{}
		if err := decoder.Decode(&rawCard); err != nil {
			log.Fatalf("failed to decode card: %w", err)
			return fmt.Errorf("failed to decode card: %w", err)
		}

		card := models.MapScryfallToCard(rawCard)
		cards = append(cards, card)

		// When batch is full, insert
		if len(cards) >= batchSize {
			if err := batchInsertCards(cards); err != nil {
				log.Fatalf("failed to insert batch: %w", err)
				return fmt.Errorf("failed to insert batch: %w", err)
			}

			totalCount += len(cards)
			batchCount++

			// Progress update every 10 batches (10,000 cards)
			if batchCount%10 == 0 {
				elapsed := time.Since(startTime)
				rate := float64(totalCount) / elapsed.Seconds()
				log.Printf("Inserted %d cards (%.0f cards/sec)...\n", totalCount, rate)
			}

			cards = cards[:0] // Reset slice
		}
	}

	// Insert remaining cards
	if len(cards) > 0 {
		if err := batchInsertCards(cards); err != nil {
			log.Fatalf("failed to insert final batch: %w", err)
			return fmt.Errorf("failed to insert final batch: %w", err)
		}
		totalCount += len(cards)
	}

	elapsed := time.Since(startTime)
	log.Printf("\nPriming complete! Inserted %d cards in %s (%.0f cards/sec)\n",
		totalCount, elapsed.Round(time.Second), float64(totalCount)/elapsed.Seconds())
	return nil
}
