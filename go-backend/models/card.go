package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

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


// MapScryfallToCard converts Scryfall JSON to Card models
func MapScryfallToCard(data map[string]interface{}) *Card {
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