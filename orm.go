package firegorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// BaseModel defines the core structure and behavior for Firestore models.
type BaseModel struct {
	ID            string    `firestore:"id" json:"id"`
	CreatedAt     time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt     time.Time `firestore:"updated_at" json:"updated_at"`
	Deleted       bool      `firestore:"deleted" json:"deleted"`
	CollectionName string   `firestore:"-" json:"-"` // Not persisted in Firestore
}

// SetCollectionName explicitly sets the collection name.
func (b *BaseModel) SetCollectionName(name string) {
	b.CollectionName = name
}

// EnsureCollection ensures that the collection name is set.
func (b *BaseModel) EnsureCollection() error {
	if b.CollectionName == "" {
		return errors.New("collection name not set; ensure the model is properly initialized")
	}
	return nil
}

// Create inserts a new document into the model's collection.
func (b *BaseModel) Create(ctx context.Context, data interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		return err
	}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to a struct")
	}

	// Set ID and timestamps
	b.setID(generateUUID())
	b.setTimestamps()

	val = val.Elem()
	val.FieldByName("ID").SetString(b.ID)
	val.FieldByName("CreatedAt").Set(reflect.ValueOf(b.CreatedAt))
	val.FieldByName("UpdatedAt").Set(reflect.ValueOf(b.UpdatedAt))

	log.Printf("Creating document in collection '%s': %+v", b.CollectionName, data)
	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Set(ctx, data)
	return err
}

// Get retrieves a document by ID and maps it to the provided model.
func (b *BaseModel) Get(ctx context.Context, id string, model interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		return err
	}

	doc, err := Client.Collection(b.CollectionName).Doc(id).Get(ctx)
	if err != nil {
		return err
	}

	if err := doc.DataTo(model); err != nil {
		return err
	}

	log.Printf("Fetched document from collection '%s': %+v", b.CollectionName, model)
	return nil
}

// Update modifies specific fields of a document.
func (b *BaseModel) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		return err
	}

	updates["updated_at"] = firestore.ServerTimestamp
	log.Printf("Updating document ID '%s' in collection '%s' with updates: %+v", id, b.CollectionName, updates)
	_, err := Client.Collection(b.CollectionName).Doc(id).Update(ctx, updatesToFirestoreUpdates(updates))
	return err
}

// Delete performs a soft delete by marking the document as deleted.
func (b *BaseModel) Delete(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"deleted":    true,
		"updated_at": firestore.ServerTimestamp,
	}
	return b.Update(ctx, id, updates)
}

// List retrieves documents with optional filters and maps them to the provided results slice.
func (b *BaseModel) List(ctx context.Context, filters map[string]interface{}, limit int, startAfter string, results interface{}) (string, error) {
	if err := b.EnsureCollection(); err != nil {
		return "", err
	}

	query := Client.Collection(b.CollectionName).Where("deleted", "==", false)
	for field, value := range filters {
		query = query.Where(field, "==", value)
	}

	if startAfter != "" {
		doc, err := Client.Collection(b.CollectionName).Doc(startAfter).Get(ctx)
		if err != nil {
			return "", fmt.Errorf("invalid startAfter token: %v", err)
		}
		query = query.StartAfter(doc)
	}

	iter := query.Limit(limit).Documents(ctx)
	defer iter.Stop()

	resultsVal := reflect.ValueOf(results).Elem()
	for {
		doc, err := iter.Next()
		if err == iterator.Done { // Corrected iterator.Done usage
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to iterate documents: %v", err)
		}

		item := reflect.New(resultsVal.Type().Elem()).Interface()
		if err := doc.DataTo(item); err != nil {
			return "", fmt.Errorf("failed to map document data: %v", err)
		}

		resultsVal.Set(reflect.Append(resultsVal, reflect.ValueOf(item)))
	}

	nextPageToken := ""
	if resultsVal.Len() == limit {
		lastItem := resultsVal.Index(resultsVal.Len() - 1).Interface().(*BaseModel)
		nextPageToken = lastItem.ID
	}

	log.Printf("Listed documents from collection '%s': %+v", b.CollectionName, results)
	return nextPageToken, nil
}

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
