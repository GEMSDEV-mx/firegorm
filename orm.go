package firegorm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Create inserts a new document into the model's collection.
func (b *BaseModel) Create(ctx context.Context, data interface{}) error {
	if err := b.EnsureCollection(data); err != nil {
		return err
	}

	// Set ID and timestamps
	b.setID(generateUUID())
	b.setTimestamps()

	// Reflectively update the incoming data with the BaseModel fields
	val := reflect.ValueOf(data).Elem()
	val.FieldByName("ID").SetString(b.ID)
	val.FieldByName("CreatedAt").Set(reflect.ValueOf(b.CreatedAt))
	val.FieldByName("UpdatedAt").Set(reflect.ValueOf(b.UpdatedAt))

	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Set(ctx, data)
	return err
}


// Get retrieves a document by ID and maps it to the registered model schema.
func (b *BaseModel) Get(ctx context.Context, id string, model interface{}) error {
	if err := b.EnsureCollection(model); err != nil {
		return err
	}

	doc, err := Client.Collection(b.CollectionName).Doc(id).Get(ctx)
	if err != nil {
		return err
	}

	// Map Firestore data to the provided model instance
	if err := doc.DataTo(model); err != nil {
		return err
	}

	log.Printf("Fetched document from collection '%s': %+v", b.CollectionName, model)
	return nil
}

// Update modifies specific fields of a document.
func (b *BaseModel) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if b.CollectionName == "" {
		return errors.New("collection name not set; ensure the collection is properly initialized")
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

// List retrieves documents with optional filters and maps them to the registered schema.
func (b *BaseModel) List(ctx context.Context, filters map[string]interface{}, limit int, startAfter string, results interface{}) (string, error) {
	if b.CollectionName == "" {
		return "", errors.New("collection name not set; ensure the collection is properly initialized")
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
		if err == iterator.Done {
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
