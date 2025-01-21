package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadCredentials() string {
    err := godotenv.Load()
    if err != nil {
        log.Println("Error loading .env file")
    }
	credentials := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if credentials == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_KEY is not set")
	}
	return credentials
}
