package firegorm

import (
	"errors"
	"time"
)

// BaseModel defines the core structure and behavior for Firestore models.
type BaseModel struct {
	ID            string     `firestore:"id" json:"id"`
	CreatedAt     time.Time  `firestore:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `firestore:"updated_at" json:"updated_at"`
	Deleted       bool       `firestore:"deleted" json:"deleted"`
	DeletedAt     *time.Time `firestore:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	CollectionName string    `firestore:"-" json:"-"` // Not persisted in Firestore
	ModelName      string    `firestore:"-" json:"-"` // Not persisted in Firestore
}

// SetCollectionName explicitly sets the collection name.
func (b *BaseModel) SetCollectionName(name string) {
	b.CollectionName = name
	Log(DEBUG, "Set collection name to '%s'", name)
}

// SetModelName explicitly sets the model name.
func (b *BaseModel) SetModelName(name string) {
	b.ModelName = name
	Log(DEBUG, "Set model name to '%s'", name)
}

// GetCollectionName returns the collection name.
func (b *BaseModel) GetCollectionName() string {
	Log(DEBUG, "Getting collection name: '%s'", b.CollectionName)
	return b.CollectionName
}

// GetModelName returns the model name.
func (b *BaseModel) GetModelName() string {
	Log(DEBUG, "Getting model name: '%s'", b.ModelName)
	return b.ModelName
}

// EnsureCollection ensures that the collection name is set.
func (b *BaseModel) EnsureCollection() error {
	if b.CollectionName == "" {
		Log(WARN, "Collection name is not set in BaseModel")
		return errors.New("collection name not set; ensure the model is properly initialized")
	}
	Log(DEBUG, "Collection name '%s' is properly set", b.CollectionName)
	return nil
}
