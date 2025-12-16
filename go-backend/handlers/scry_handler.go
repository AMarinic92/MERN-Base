package handlers

import (
	"encoding/json"
	"fmt"
	"go-backend/database"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func SearchCard(w http.ResponseWriter, r *http.Request) {
	// Get card name from query parameter
	cardName := r.URL.Query().Get("name")
	if cardName == "" {
		http.Error(w, "Card name is required", http.StatusBadRequest)
		return
	}

	// First, check the database
	card, err := database.SearchCardByName(cardName)
	if err == nil {
		// Card found in database
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"source": "cache",
			"card":   card,
		})
		return
	}

	// If not found in database (or error other than not found), fetch from Scryfall
	if err != gorm.ErrRecordNotFound {
		// Log unexpected database error but continue to Scryfall
		fmt.Printf("Database error: %v\n", err)
	}

	// Fetch from Scryfall API
	scryfallURL := fmt.Sprintf("https://api.scryfall.com/cards/named?exact=%s", cardName)
	resp, err := http.Get(scryfallURL)
	if err != nil {
		http.Error(w, "Failed to fetch from Scryfall", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Card not found"+strconv.Itoa(resp.StatusCode), http.StatusNotFound)
		return
	}

	// Read Scryfall response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read Scryfall response", http.StatusInternalServerError)
		return
	}

	// Parse Scryfall response
	var scryfallCard map[string]interface{}
	if err := json.Unmarshal(body, &scryfallCard); err != nil {
		http.Error(w, "Failed to parse Scryfall response", http.StatusInternalServerError)
		return
	}

	// Cache the card in database
	newCard := mapScryfallToCard(scryfallCard)
	if err := database.UpsertCard(newCard); err != nil {
		fmt.Printf("Failed to cache card: %v\n", err)
		// Continue anyway - we still have the data to return
	}

	// Return the card from Scryfall
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"source": "scryfall",
		"card":   newCard,
	})
}

func GetRndCard(w http.ResponseWriter, r *http.Request) {

	// Check if card already exists in database
	existingCard, err := database.GetRandomCard()
	if err == nil {
		// Card found in database, return cached version
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"card": existingCard,
		})
		return
	}

}

func GetSimilarCards(w http.ResponseWriter, r *http.Request) {
	// Read the raw body
	log.Println("=== GetSimilarCards HANDLER CALLED ===")
	log.Printf("Method: %s, Path: %s", r.Method, r.URL.Path)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// If you're expecting JSON, unmarshal it into a struct
	var requestData struct {
		Name        string   `json:"name"`
		OracleTexts []string `json:"oracle_texts"`
	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use the data
	cards, err := database.SearchFuzzyOracleText(requestData.Name, requestData.OracleTexts)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// mapScryfallToCard converts Scryfall JSON to our Card model
func mapScryfallToCard(data map[string]interface{}) *database.Card {
	card := &database.Card{}

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

	// Arrays - Updated to use pq.StringArray
	if colors, ok := data["colors"].([]interface{}); ok {
		card.Colors = make(pq.StringArray, 0, len(colors))
		for _, c := range colors {
			if color, ok := c.(string); ok {
				card.Colors = append(card.Colors, color)
			}
		}
	}

	if colorIdentity, ok := data["color_identity"].([]interface{}); ok {
		card.ColorIdentity = make(pq.StringArray, 0, len(colorIdentity))
		for _, c := range colorIdentity {
			if color, ok := c.(string); ok {
				card.ColorIdentity = append(card.ColorIdentity, color)
			}
		}
	}

	// Handle keywords if present in the data
	if keywords, ok := data["keywords"].([]interface{}); ok {
		card.Keywords = make(pq.StringArray, 0, len(keywords))
		for _, k := range keywords {
			if keyword, ok := k.(string); ok {
				card.Keywords = append(card.Keywords, keyword)
			}
		}
	}

	// Store complex objects as JSON strings
	if imageURIs, ok := data["image_uris"].(map[string]interface{}); ok {
		if jsonBytes, err := json.Marshal(imageURIs); err == nil {
			jsonStr := string(jsonBytes)
			card.ImageURIs = &jsonStr
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

	return card
}
