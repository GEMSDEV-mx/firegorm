package firegorm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// ValidateStruct validates the struct fields based on tags.
func ValidateStruct(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		Log(ERROR, "Validation failed: data must be a struct or pointer to a struct")
		return errors.New("data must be a struct or a pointer to a struct")
	}

	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("validate")
		value := val.Field(i)

		// Check for "required" tag
		if tag == "required" && value.IsZero() {
			err := fmt.Errorf("field '%s' is required", field.Name)
			Log(ERROR, "Validation failed: %v", err)
			return err
		}
	}

	Log(DEBUG, "Struct validation passed for %+v", data)
	return nil
}

// Create inserts a new document into the model's collection.
func (b *BaseModel) Create(ctx context.Context, data interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "Create failed: %v", err)
		return err
	}

	if err := ValidateStruct(data); err != nil {
		return err
	}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		err := fmt.Errorf("data must be a pointer to a struct")
		Log(ERROR, "Create failed: %v", err)
		return err
	}

	// Set ID and timestamps
	b.setID(generateUUID())
	b.setTimestamps()

	val = val.Elem()
	val.FieldByName("ID").SetString(b.ID)
	val.FieldByName("CreatedAt").Set(reflect.ValueOf(b.CreatedAt))
	val.FieldByName("UpdatedAt").Set(reflect.ValueOf(b.UpdatedAt))
	val.FieldByName("Deleted").SetBool(false)

	Log(INFO, "Creating document in collection '%s': %+v", b.CollectionName, data)
	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Set(ctx, data)
	return err
}

// Get retrieves a document by ID and maps it to the provided model.
func (b *BaseModel) Get(ctx context.Context, id string, model interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "Get failed: %v", err)
		return err
	}

	doc, err := Client.Collection(b.CollectionName).Doc(id).Get(ctx)
	if err != nil {
		Log(ERROR, "Failed to fetch document with ID '%s' from collection '%s': %v", id, b.CollectionName, err)
		return err
	}

	if err := doc.DataTo(model); err != nil {
		Log(ERROR, "Failed to map document data to model: %v", err)
		return err
	}

	Log(INFO, "Fetched document from collection '%s': %+v", b.CollectionName, model)
	return nil
}

// Update modifies specific fields of a document.
func (b *BaseModel) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "Update failed: %v", err)
		return err
	}

	// Validate updates using the registry
	if err := validateUpdateFields(updates, b); err != nil {
		Log(ERROR, "Update failed: %v", err)
		return err
	}

	// Add Firestore timestamp
	updates["updated_at"] = firestore.ServerTimestamp
	Log(INFO, "Updating document ID '%s' in collection '%s' with updates: %+v", id, b.CollectionName, updates)

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
        Log(ERROR, "List failed: %v", err)
        return "", err
    }

    query := Client.Collection(b.CollectionName).Where("deleted", "==", false)
    for field, value := range filters {
        query = query.Where(field, "==", value)
    }

    if startAfter != "" {
        doc, err := Client.Collection(b.CollectionName).Doc(startAfter).Get(ctx)
        if err != nil {
            err = fmt.Errorf("invalid startAfter token: %v", err)
            Log(ERROR, "List failed: %v", err)
            return "", err
        }
        query = query.StartAfter(doc)
    }

    iter := query.Limit(limit).Documents(ctx)
    defer iter.Stop()

    resultsVal := reflect.ValueOf(results).Elem()
    itemType := resultsVal.Type().Elem()

    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            Log(ERROR, "Failed to iterate documents: %v", err)
            return "", fmt.Errorf("failed to iterate documents: %v", err)
        }

        // Create a new item and map Firestore document data to it
        item := reflect.New(itemType).Interface()
        if err := doc.DataTo(item); err != nil {
            Log(ERROR, "Failed to map document data: %v", err)
            return "", fmt.Errorf("failed to map document data: %v", err)
        }

        // If the slice holds non-pointer values, dereference the item before appending
        if itemType.Kind() != reflect.Ptr {
            item = reflect.ValueOf(item).Elem().Interface()
        }

        resultsVal.Set(reflect.Append(resultsVal, reflect.ValueOf(item)))
    }

    nextPageToken := ""
    if resultsVal.Len() == limit {
        lastItem := resultsVal.Index(resultsVal.Len() - 1).Interface()
        nextPageToken = reflect.ValueOf(lastItem).FieldByName("ID").String()
    }

    Log(INFO, "Listed documents from collection '%s': %+v", b.CollectionName, results)
    return nextPageToken, nil
}


func (b *BaseModel) setID(id string) {
	b.ID = id
	Log(DEBUG, "Set ID for model: %s", id)
}

func (b *BaseModel) setTimestamps() {
	now := time.Now()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	Log(DEBUG, "Set timestamps: CreatedAt=%v, UpdatedAt=%v", b.CreatedAt, b.UpdatedAt)
}

// validateUpdateFields validates the fields being updated based on the struct's tags.
func validateUpdateFields(updates map[string]interface{}, baseModel *BaseModel) error {
	collectionName := baseModel.CollectionName
	modelName := baseModel.ModelName

	// Retrieve metadata from the registry
	modelInfo, err := GetModelInfo(collectionName + "." + modelName)
	if err != nil {
		Log(ERROR, "Validation failed: %v", err)
		return fmt.Errorf("collection '%s' is not registered", collectionName)
	}

	// Validate updates using the tag-to-field map
	for updateKey, value := range updates {
		fieldName, exists := modelInfo.TagToFieldMap[updateKey]
		if !exists {
			err = fmt.Errorf("field '%s' does not exist in the model for collection '%s'", updateKey, collectionName)
			Log(ERROR, "%v", err)
			return err
		}

		// Validate required fields
		field, _ := modelInfo.Schema.FieldByName(fieldName)
		validateTag := field.Tag.Get("validate")
		if validateTag == "required" && (value == nil || (reflect.ValueOf(value).Kind() == reflect.String && value == "")) {
			err = fmt.Errorf("field '%s' is required and cannot be empty or nil", updateKey)
			Log(ERROR, "%v", err)
			return err
		}
	}

	Log(DEBUG, "Validation passed for updates: %+v", updates)
	return nil
}
