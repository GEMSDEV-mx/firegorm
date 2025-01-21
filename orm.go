package firegorm

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Create inserts a new document into the model's collection.
func (b *BaseModel) Create(ctx context.Context) error {
	if err := b.ensureCollection(); err != nil {
		return err
	}

	b.setID(generateUUID())
	b.setTimestamps()

	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Set(ctx, b)
	return err
}

// Get retrieves a document by ID.
func (b *BaseModel) Get(ctx context.Context, id string) error {
	if err := b.ensureCollection(); err != nil {
		return err
	}

	doc, err := Client.Collection(b.CollectionName).Doc(id).Get(ctx)
	if err != nil {
		return err
	}

	if err := doc.DataTo(b); err != nil {
		return err
	}

	if b.Deleted {
		return fmt.Errorf("document with ID %s is deleted", id)
	}

	return nil
}

// Update modifies specific fields of the model.
func (b *BaseModel) Update(ctx context.Context, updates map[string]interface{}) error {
	if err := b.ensureCollection(); err != nil {
		return err
	}

	updates["updated_at"] = firestore.ServerTimestamp
	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Update(ctx, updatesToFirestoreUpdates(updates))
	return err
}

// Delete performs a soft delete by marking the model as deleted.
func (b *BaseModel) Delete(ctx context.Context) error {
	if err := b.ensureCollection(); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"deleted":    true,
		"updated_at": firestore.ServerTimestamp,
	}
	return b.Update(ctx, updates)
}

// List retrieves documents with optional filters.
func (b *BaseModel) List(ctx context.Context, filters map[string]interface{}, limit int, startAfter string) ([]map[string]interface{}, string, error) {
	if err := b.ensureCollection(); err != nil {
		return nil, "", err
	}

	query := Client.Collection(b.CollectionName).Where("deleted", "==", false)

	for field, value := range filters {
		query = query.Where(field, "==", value)
	}

	if startAfter != "" {
		doc, err := Client.Collection(b.CollectionName).Doc(startAfter).Get(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("invalid startAfter token: %v", err)
		}
		query = query.StartAfter(doc)
	}

	iter := query.Limit(limit).Documents(ctx)
	defer iter.Stop()

	var results []map[string]interface{}
	var lastDocID string

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to iterate documents: %v", err)
		}

		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
		lastDocID = doc.Ref.ID
	}

	nextPageToken := ""
	if len(results) == limit {
		nextPageToken = lastDocID
	}

	return results, nextPageToken, nil
}
