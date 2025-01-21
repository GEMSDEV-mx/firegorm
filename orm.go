package firegorm

import (
	"context"
	"fmt"
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
func (b *BaseModel) Get(ctx context.Context, id string, modelName string) (interface{}, error) {
	info, err := GetModelInfo(modelName)
	if err != nil {
		return nil, err
	}

	doc, err := Client.Collection(info.CollectionName).Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the model schema
	data := reflect.New(info.Schema).Interface()
	if err := doc.DataTo(data); err != nil {
		return nil, err
	}

	return data, nil
}

// Update modifies specific fields of a document.
func (b *BaseModel) Update(ctx context.Context, id string, updates map[string]interface{}, modelName string) error {
	info, err := GetModelInfo(modelName)
	if err != nil {
		return err
	}

	// Update the timestamp
	updates["updated_at"] = firestore.ServerTimestamp

	_, err = Client.Collection(info.CollectionName).Doc(id).Update(ctx, updatesToFirestoreUpdates(updates))
	return err
}


// Delete performs a soft delete by marking the document as deleted.
func (b *BaseModel) Delete(ctx context.Context, id string, modelName string) error {
	updates := map[string]interface{}{
		"deleted":    true,
		"updated_at": firestore.ServerTimestamp,
	}
	return b.Update(ctx, id, updates, modelName)
}

// List retrieves documents with optional filters and maps them to the registered schema.
func (b *BaseModel) List(ctx context.Context, filters map[string]interface{}, limit int, startAfter string, modelName string) ([]interface{}, string, error) {
	info, err := GetModelInfo(modelName)
	if err != nil {
		return nil, "", err
	}

	query := Client.Collection(info.CollectionName).Where("deleted", "==", false)
	for field, value := range filters {
		query = query.Where(field, "==", value)
	}

	if startAfter != "" {
		doc, err := Client.Collection(info.CollectionName).Doc(startAfter).Get(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("invalid startAfter token: %v", err)
		}
		query = query.StartAfter(doc)
	}

	iter := query.Limit(limit).Documents(ctx)
	defer iter.Stop()

	var results []interface{}
	var lastDocID string

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to iterate documents: %v", err)
		}

		data := reflect.New(info.Schema).Interface()
		if err := doc.DataTo(data); err != nil {
			return nil, "", fmt.Errorf("failed to parse document data: %v", err)
		}

		results = append(results, data)
		lastDocID = doc.Ref.ID
	}

	nextPageToken := ""
	if len(results) == limit {
		nextPageToken = lastDocID
	}

	return results, nextPageToken, nil
}
