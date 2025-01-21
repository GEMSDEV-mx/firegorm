package firegorm

import (
	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// Generate UUID for document IDs.
func generateUUID() string {
	id := uuid.New().String()
	Log(DEBUG, "Generated UUID: %s", id)
	return id
}

// Convert map to Firestore updates.
func updatesToFirestoreUpdates(updates map[string]interface{}) []firestore.Update {
	var firestoreUpdates []firestore.Update
	for key, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
		Log(DEBUG, "Added Firestore update: Path=%s, Value=%v", key, value)
	}
	Log(INFO, "Converted updates to Firestore format: %+v", firestoreUpdates)
	return firestoreUpdates
}
