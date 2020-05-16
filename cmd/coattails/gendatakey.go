package main

import (
	"encoding/hex"
	"log"

	"github.com/bluedresscapital/coattails/pkg/secrets"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	secrets.InitSecret()
	// Generate datakey
	dataKey := uuid.New().String()
	log.Printf("Generated the following datakey: %s. NEVER PERSISTENTLY STORE THIS", dataKey)
	// Encrypt it, and output its encrypted value
	cipher := secrets.Encrypt(dataKey)
	cipherStr := hex.EncodeToString(cipher)
	log.Printf("Here is your encrypted datakey: %s. Store it somewhere safe, AND NEVER LOSE IT", cipherStr)
}
