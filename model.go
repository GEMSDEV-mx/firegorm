package firegorm

import (
	"reflect"
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

// EnsureCollection ensures that the model has a valid collection name.
func (b *BaseModel) EnsureCollection(model interface{}) error {
	if b.CollectionName == "" {
		modelName := reflect.TypeOf(model).Name()
		info, err := GetModelInfo(modelName)
		if err != nil {
			return err
		}
		b.CollectionName = info.CollectionName
	}
	return nil
}

// setID sets the ID for the model.
func (b *BaseModel) setID(id string) {
	b.ID = id
}

// setTimestamps sets the CreatedAt and UpdatedAt timestamps.
func (b *BaseModel) setTimestamps() {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
}
