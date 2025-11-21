package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"go-backend/database"
	"go-backend/models"
)

// GetProducts handles GET /products to retrieve all products.
func GetProducts(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var products []models.Product
	// Query the database using the shared DB instance
	if result := database.DB.Find(&products); result.Error != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

// CreateProduct handles POST /products to add a new product.
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if result := database.DB.Create(&product); result.Error != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// DeleteProduct handles DELETE /products/{id} to remove a product.
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	vars := mux.Vars(r)
	id := vars["id"]

	// Delete the product by ID
	if result := database.DB.Delete(&models.Product{}, id); result.Error != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Product successfully deleted"})
}

// OptionsHandler handles CORS preflight requests
func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
}
