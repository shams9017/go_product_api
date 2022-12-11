package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"

    _ "github.com/lib/pq"
)

http.HandleFunc("/product", getProduct).Methods("GET")
http.HandleFunc("/products", getProducts).Methods("GET")
http.HandleFunc("/product", deleteProduct).Methods("DELETE")
http.HandleFunc("/product", updateProduct).Methods("PUT")

// Start the server.
log.Fatal(http.ListenAndServe(":8080", nil))

// Product represents a product in the database.
type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Price       float64 `json:"price"`
}

// Products is a collection of Product objects.
type Products []Product

// DB is a global variable that represents the database connection.
var DB *sql.DB

// ErrorResponse is a helper struct for returning error messages in a standard format.
type ErrorResponse struct {
    Error string `json:"error"`
}

// getProduct retrieves a single product from the database based on the product ID.
func getProduct(w http.ResponseWriter, r *http.Request) {
    // Get the product ID from the URL query string.
    productIDStr := r.URL.Query().Get("id")
    productID, err := strconv.Atoi(productIDStr)
    if err != nil {
        // If the product ID is not a valid integer, return an error.
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid product ID."})
        return
    }

    // Query the database for the product with the given ID.
    row := DB.QueryRow("SELECT id, name, category, price FROM products WHERE id = $1", productID)

    // Scan the result into a Product object.
    var product Product
    err = row.Scan(&product.ID, &product.Name, &product.Category, &product.Price)
    if err == sql.ErrNoRows {
        // If there is no product with the given ID, return an error.
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Product not found."})
        return
    } else if err != nil {
        // If there is any other error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve product."})
        return
    }

    // If everything went well, return the product in the response body.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// getProducts retrieves a list of products from the database based on the query parameters.
// getProducts retrieves a list of products from the database based on the query parameters.
func getProducts(w http.ResponseWriter, r *http.Request) {
    // Parse the query parameters into a map.
    queryValues := r.URL.Query()

    // Build the WHERE clause of the SQL query based on the query parameters.
    var whereClauses []string
    var whereArgs []interface{}
    if name := queryValues.Get("name"); name != "" {
        whereClauses = append(whereClauses, "name LIKE $1")
        whereArgs = append(whereArgs, fmt.Sprintf("%%%s%%", name))
    }
    if category := queryValues.Get("category"); category != "" {
        whereClauses = append(whereClauses, "category = $2")
        whereArgs = append(whereArgs, category)
    }
    if minPriceStr := queryValues.Get("min_price"); minPriceStr != "" {
        minPrice, err := strconv.ParseFloat(minPriceStr, 64)
        if err != nil {
            // If the minimum price is not a valid float, return an error.
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid minimum price."})
            return
        }
        whereClauses = append(whereClauses, "price >= $3")
        whereArgs = append(whereArgs, minPrice)
    }
    if maxPriceStr := queryValues.Get("max_price"); maxPriceStr != "" {
        maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
        if err != nil {
            // If the maximum price is not a valid float, return an error.
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid maximum price."})
            return
        }
        whereClauses = append(whereCla

uses, "price <= $4")
        whereArgs = append(whereArgs, maxPrice)
    }

    // Build the final SQL query.
    query := "SELECT id, name, category, price FROM products"
    if len(whereClauses) > 0 {
        query += fmt.Sprintf(" WHERE %s", strings.Join(whereClauses, " AND "))
    }

    // Query the database for the products that match the WHERE clause.
    rows, err := DB.Query(query, whereArgs...)
    if err != nil {
        // If there is an error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to retrieve products."})
        return
    }
    defer rows.Close()

    // Scan the results into a slice of Product objects.
    var products Products
    for rows.Next() {
        var product Product
        err := rows.Scan(&product.ID, &product.Name, &product.Category, &product.Price)
		if err != nil {
			// If the product ID is not a valid integer, return an error.
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid product ID."})
			return
		}
	}



// deleteProduct deletes a single product from the database based on the product ID.
func deleteProduct(w http.ResponseWriter, r *http.Request) {
    // Get the product ID from the URL query string.
    productIDStr := r.URL.Query().Get("id")
    productID, err := strconv.Atoi(productIDStr)
    if err != nil {
        // If the product ID is not a valid integer, return an error.
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid product ID."})
        return
    }

    // Delete the product with the given ID from the database.
    result, err := DB.Exec("DELETE FROM products WHERE id = $1", productID)
    if err != nil {
        // If there is an error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete product."})
        return
    }

    // Get the number of rows affected by the DELETE operation.
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        // If there is an error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete product."})
        return
    }
    if rowsAffected == 0 {
        // If no rows were affected by the DELETE operation, the product with the given ID
        // must not exist in the database, so return a 404 Not Found response.
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Product not found."})
        return
    }

    // If everything went well, return a 204 No Content response.
    w.WriteHeader(http.StatusNoContent)
}

// updateProduct updates a single product in the database based on the product ID.
func updateProduct(w http.ResponseWriter, r *http.Request) {
    // Get the product ID from the URL query string.
    productIDStr := r.URL.Query().Get("id")
    productID, err := strconv.Atoi(productIDStr)
    if err != nil {
        // If the product ID is not a valid integer, return an error.
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid product ID."})
        return
    }

    // Read the request body into a Product object.
    var product Product
    err = json.NewDecoder(r.Body).Decode(&product)
    if err != nil {
        // If there is an error, log it and return a 400 Bad Request response.
        log.Println(err)
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to parse request body."})
        return
    }

    // Update the product with the given ID in the database.
    result, err := DB.Exec("UPDATE products SET name = $1, category = $2, price = $3 WHERE id = $4",
        product.Name, product.Category, product.Price, productID)
    if err != nil {
        // If there is an error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update product."})
        return
    }

    // Get the number of rows affected by the UPDATE operation.
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        // If there is an error, log it and return a 500 Internal Server Error response.
        log.Println(err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update product."})
        return
    }
    if rowsAffected == 0 {
        // If no rows were affected by the UPDATE operation, the product with the given ID
        // must not exist in the database, so return a 404 Not Found response.
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(ErrorResponse{Error: "Product not found."})
        return
    }

    // If everything went well, return the updated product in the response body.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

