package firegorm

import (
	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// Generate UUID for document IDs.
func generateUUID() string {
	return uuid.New().String()
}

// Convert map to Firestore updates.
func updatesToFirestoreUpdates(updates map[string]interface{}) []firestore.Update {
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}
	return firestoreUpdates
}
