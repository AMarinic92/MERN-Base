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

// UpdateProduct handles PUT /products/{id} to update an existing product.
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    vars := mux.Vars(r)
    id := vars["id"]

    var existingProduct models.Product
    // 1. Check if the product exists
    if result := database.DB.First(&existingProduct, id); result.Error != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    var updatedData models.Product
    // 2. Decode the incoming JSON payload
    if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // 3. Update the fields and save
    // Using Map allows selective updating based on what was provided in the JSON body.
    if result := database.DB.Model(&existingProduct).Updates(updatedData); result.Error != nil {
        http.Error(w, "Failed to update product", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(existingProduct)
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