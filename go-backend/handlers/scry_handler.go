package handlers

import (
	"encoding/json"
	"go-backend/database"
	"go-backend/models"
	"io"
	"net/http"

	"github.com/lib/pq"
)

func GetCardID(w http.ResponseWriter, r *http.Request){
		cardId := r.URL.Query().Get("id")
	if cardId == "" {
		http.Error(w, "Card name is required", http.StatusBadRequest)
		return
	}
		card, err := database.GetCardByID(cardId)
	if err == nil {
		// Card found in database, return cached version
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": card,
		})
		return
	}
}


func GetRndCard(w http.ResponseWriter, r *http.Request) {

	// Check if card already exists in database
	card, err := database.GetRandomCard()
	if err == nil {
		// Card found in database, return cached version
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"card": card,
		})
		return
	}

}

func GetSimilarCards(w http.ResponseWriter, r *http.Request) {
	// Read the raw body
	
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

func MemSuggest(w http.ResponseWriter, r *http.Request){
		// Read the raw body
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// If you're expecting JSON, unmarshal it into a struct
	var requestData struct {
		OracleID        string   `json:"oracle_id"`

	}

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Use the data
	cards, err := database.GetCardSuggestions(requestData.OracleID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

func GetFuzzyCard(w http.ResponseWriter, r *http.Request){
	cardName := r.URL.Query().Get("name")
	if cardName == "" {
		http.Error(w, "Card name is required", http.StatusBadRequest)
		return
	}
		card, err := database.SearchCardByNameFuzzy(cardName)
	if err == nil {
		// Card found in database, return cached version
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": card,
		})
		return
	}

}

// mapScryfallToCard converts Scryfall JSON to our Card model
func mapScryfallToCard(data map[string]interface{}) *models.Card {
	card := &models.Card{}

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
func OptionsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    w.WriteHeader(http.StatusOK)
}