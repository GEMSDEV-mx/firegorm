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

// EnsureCollection ensures that the collection name is set.

// ValidateStruct validates the struct fields based on tags.
func ValidateStruct(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("data must be a struct or a pointer to a struct")
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("validate")
		value := val.Field(i)

		// Check for "required" tag
		if tag == "required" && value.IsZero() {
			return fmt.Errorf("field '%s' is required", field.Name)
		}

		// Additional validation rules can be added here
	}

	return nil
}

// Create inserts a new document into the model's collection.
func (b *BaseModel) Create(ctx context.Context, data interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		return err
	}

	if err := ValidateStruct(data); err != nil {
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
	val.FieldByName("Deleted").SetBool(false)

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

	// Validate updates using the registry
	if err := validateUpdateFields(updates, b); err != nil {
		return err
	}

	// Add Firestore timestamp
	updates["updated_at"] = firestore.ServerTimestamp
	log.Printf("Updating document ID '%s' in collection '%s' with updates: %+v", id, b.CollectionName, updates)

	_, err := Client.Collection(b.CollectionName).Doc(id).Update(ctx, updatesToFirestoreUpdates(updates))
	return err
}

// Delete performs a soft delete by marking the document as deleted.
func (b *BaseModel) Delete(ctx context.Context, id string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"deleted":    true,
		"deleted_at": now,
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

// validateUpdateFields validates the fields being updated based on the struct's tags.
func validateUpdateFields(updates map[string]interface{}, baseModel *BaseModel) error {
	// Use the CollectionName from the base model
	collectionName := baseModel.CollectionName

	// Retrieve metadata from the registry
	modelInfo, err := GetModelInfo(collectionName)
	if err != nil {
		return fmt.Errorf("collection '%s' is not registered", collectionName)
	}

	// Validate updates using the tag-to-field map
	for updateKey, value := range updates {
		fieldName, exists := modelInfo.TagToFieldMap[updateKey]
		if !exists {
			return fmt.Errorf("field '%s' does not exist in the model for collection '%s'", updateKey, collectionName)
		}

		// Validate required fields
		field, _ := modelInfo.Schema.FieldByName(fieldName)
		validateTag := field.Tag.Get("validate")
		if validateTag == "required" && (value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value == "")) {
			return fmt.Errorf("field '%s' is required and cannot be empty or nil", updateKey)
		}
	}

	return nil
}
