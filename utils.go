package firegorm

import (
	"strings"

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

// ExtractFilters converts a map of query parameters (key-value strings) into a filters map.
// Any parameter value that contains a comma is split into a []string.
// The caller can pass a slice of keys to exclude from processing.
func ExtractFilters(params map[string]string, exclude []string) map[string]interface{} {
    filters := make(map[string]interface{})
    // Build an exclusion lookup.
    excludeMap := make(map[string]bool)
    for _, key := range exclude {
        excludeMap[key] = true
    }
    for key, value := range params {
        if excludeMap[key] {
            continue
        }
        if strings.Contains(value, ",") {
            parts := strings.Split(value, ",")
            for i, part := range parts {
                parts[i] = strings.TrimSpace(part)
            }
            filters[key] = parts
        } else {
            filters[key] = value
        }
    }
    return filters
}