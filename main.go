package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Struct to parse the request body
type RequestData struct {
	Values []float64 `json:"values"` // Changed to float64
}

var collection *mongo.Collection

// POST Handler to save values to MongoDB
func valuesPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var data RequestData

	// Decode the JSON body
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create a document with values and a timestamp
	document := bson.M{
		"values":    data.Values,
		"timestamp": time.Now(),
	}

	// Insert the document into MongoDB
	_, err = collection.InsertOne(context.TODO(), document)
	if err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return
	}

	// Send a response back to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Values saved successfully with timestamp"))
}

// GET Handler to retrieve values from MongoDB
func valuesGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve all documents
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		http.Error(w, "Failed to decode data", http.StatusInternalServerError)
		return
	}

	// Send the data as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	// MongoDB Atlas connection URI (replace with your connection string)
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set in env")
	}
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.TODO())

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	fmt.Println("Connected to MongoDB!")

	host := os.Getenv("SERVER_HOST")
	port := os.Getenv("SERVER_PORT")
	if host == "" || port == "" {
		log.Fatal("SERVER_HOST or SERVER_PORT not set in .env file")
	}

	// Get collection handle
	collection = client.Database("testdb").Collection("values")

	// Route setup
	http.HandleFunc("/values", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			valuesPostHandler(w, r)
		} else if r.Method == http.MethodGet {
			valuesGetHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	serverAddress := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Server running on http://%s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
