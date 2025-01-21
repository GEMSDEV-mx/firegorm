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
}

// SetModelName explicitly sets the model name.
func (b *BaseModel) SetModelName(name string) {
	b.ModelName = name
}

// GetCollectionName returns the collection name.
func (b *BaseModel) GetCollectionName() string {
	return b.CollectionName
}

// GetModelName returns the model name.
func (b *BaseModel) GetModelName() string {
	return b.ModelName
}

func (b *BaseModel) EnsureCollection() error {
	if b.CollectionName == "" {
		return errors.New("collection name not set; ensure the model is properly initialized")
	}
	return nil
}