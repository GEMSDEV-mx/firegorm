package firegorm

import (
	"errors"
	"time"
)

// BaseModel defines the core structure and behavior for Firestore models.
type BaseModel struct {
	ID            string    `firestore:"id" json:"id"`
	CreatedAt     time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt     time.Time `firestore:"updated_at" json:"updated_at"`
	Deleted       bool      `firestore:"deleted" json:"deleted"`
	CollectionName string   `firestore:"-" json:"-"` // Not persisted in Firestore
}

// Initialize the model with a collection name.
func (b *BaseModel) initModel(collectionName string) {
	b.CollectionName = collectionName
}

// Set ID and timestamps.
func (b *BaseModel) setID(id string) {
	b.ID = id
}

func (b *BaseModel) setTimestamps() {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
}

// Ensure collection name is set before any operation.
func (b *BaseModel) ensureCollection() error {
	if b.CollectionName == "" {
		return errors.New("collection name is not set")
	}
	return nil
}
