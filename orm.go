package firegorm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
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
	val.FieldByName("DeletedAt").Set(reflect.Zero(val.FieldByName("DeletedAt").Type()))
	val.FieldByName("Deleted").SetBool(false)

	Log(INFO, "Creating document in collection '%s': %+v", b.CollectionName, data)
    // after you’ve set ID & timestamps but before Set(ctx,…):
	if err := DefaultRegistry.RunHooks(ctx, b.CollectionName, PreCreate, data); err != nil {
		return err
	}
	_, err := Client.Collection(b.CollectionName).Doc(b.ID).Set(ctx, data)
	_ = DefaultRegistry.RunHooks(ctx, b.CollectionName, PostCreate, data)
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

// FindOneBy retrieves a single document from the collection that matches the given property and value.
func (b *BaseModel) FindOneBy(ctx context.Context, property string, value interface{}, model interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "FindOneBy failed: %v", err)
		return err
	}

	// Build the query: only non-deleted documents are considered.
	query := Client.Collection(b.CollectionName).
		Where(property, "==", value).
		Where("deleted", "==", false).
		Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		err = fmt.Errorf("no document found for %s == %v", property, value)
		Log(WARN, "FindOneBy: %v", err)
		return err
	}
	if err != nil {
		Log(ERROR, "Error executing query in FindOneBy: %v", err)
		return err
	}

	if err := doc.DataTo(model); err != nil {
		Log(ERROR, "Failed to map document data to model in FindOneBy: %v", err)
		return err
	}

	Log(INFO, "Found document by %s == %v in collection '%s': %+v", property, value, b.CollectionName, model)
	return nil
}

// FindOne retrieves a single document from the collection that matches the given filters.
// FindOne retrieves a single document from the collection that matches the given filters.
func (b *BaseModel) FindOne(ctx context.Context, filters map[string]interface{}, model interface{}) error {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "FindOne failed: %v", err)
		return err
	}

	// Start with a query that excludes deleted documents.
	query := Client.Collection(b.CollectionName).Where("deleted", "==", false)

	// Apply operator filters (e.g., __gt, __lte) instead of using simple equality.
	var err error
	query, err = applyOperatorFilters(query, filters)
	if err != nil {
		Log(ERROR, "FindOne failed when applying filters: %v", err)
		return err
	}

	query = query.Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		err = fmt.Errorf("no document found matching filters: %v", filters)
		Log(WARN, "FindOne: %v", err)
		return err
	}
	if err != nil {
		Log(ERROR, "Error executing query in FindOne: %v", err)
		return err
	}

	if err := doc.DataTo(model); err != nil {
		Log(ERROR, "Failed to map document data to model in FindOne: %v", err)
		return err
	}

	Log(INFO, "Found document with filters %v in collection '%s': %+v", filters, b.CollectionName, model)
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

    // — run pre-update hooks —
    if err := DefaultRegistry.RunHooks(ctx, b.CollectionName, PreUpdate, updates); err != nil {
        return err
    }

    _, err := Client.Collection(b.CollectionName).Doc(id).Update(ctx, updatesToFirestoreUpdates(updates))
    // — run post-update hooks —
    _ = DefaultRegistry.RunHooks(ctx, b.CollectionName, PostUpdate, updates)
	return err
}

// Delete performs a soft delete by marking the document as deleted.
func (b *BaseModel) Delete(ctx context.Context, id string) error {
    // — run pre-delete hooks —
    if err := DefaultRegistry.RunHooks(ctx, b.CollectionName, PreDelete, id); err != nil {
        return err
    }
	now := time.Now()
	updates := map[string]interface{}{
		"deleted":    true,
		"deleted_at": now,
		"updated_at": firestore.ServerTimestamp,
	}
    // perform the soft-delete
    err := b.Update(ctx, id, updates)
    // — run post-delete hooks —
    if err == nil {
        _ = DefaultRegistry.RunHooks(ctx, b.CollectionName, PostDelete, id)
    }
    return err	
}

// List retrieves documents with optional filters, sorting, and pagination.
func (b *BaseModel) List(ctx context.Context, filters map[string]interface{}, limit int, startAfter string, sortField string, sortOrder string, results interface{}) (string, error) {
    if err := b.EnsureCollection(); err != nil {
        Log(ERROR, "List failed: %v", err)
        return "", err
    }

    // Start with the query: only non-deleted documents.
    query := Client.Collection(b.CollectionName).Where("deleted", "==", false)
    
    // Apply operator filters (supports __gt, __gte, __lt, __lte for any field, including custom date fields)
    var err error
    query, err = applyOperatorFilters(query, filters)
    if err != nil {
        Log(ERROR, "List failed when applying filters: %v", err)
        return "", err
    }

    // Apply sorting if sortField is provided.
    if sortField != "" {
        if sortOrder == "asc" {
            query = query.OrderBy(sortField, firestore.Asc)
        } else if sortOrder == "desc" {
            query = query.OrderBy(sortField, firestore.Desc)
        } else {
            err := fmt.Errorf("invalid sortOrder: %s. Must be 'asc' or 'desc'", sortOrder)
            Log(ERROR, "List failed: %v", err)
            return "", err
        }
    }

    // If a startAfter token is provided, use it for pagination.
    if startAfter != "" {
        doc, err := Client.Collection(b.CollectionName).Doc(startAfter).Get(ctx)
        if err != nil {
            err = fmt.Errorf("invalid startAfter token: %v", err)
            Log(ERROR, "List failed: %v", err)
            return "", err
        }
        query = query.StartAfter(doc)
    }

    // Only apply limit if it's greater than zero.
    if limit > 0 {
        query = query.Limit(limit)
    }

    iter := query.Documents(ctx)
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

        // Create a new item and map Firestore document data to it.
        item := reflect.New(itemType).Interface()
        if err := doc.DataTo(item); err != nil {
            Log(ERROR, "Failed to map document data: %v", err)
            return "", fmt.Errorf("failed to map document data: %v", err)
        }

        // If the slice holds non-pointer values, dereference the item before appending.
        if itemType.Kind() != reflect.Ptr {
            item = reflect.ValueOf(item).Elem().Interface()
        }

        resultsVal.Set(reflect.Append(resultsVal, reflect.ValueOf(item)))
    }

    nextPageToken := ""
    // Only set nextPageToken if a limit was applied.
    if limit > 0 && resultsVal.Len() == limit {
        lastItem := resultsVal.Index(resultsVal.Len() - 1).Interface()
        nextPageToken = reflect.ValueOf(lastItem).FieldByName("ID").String()
    }

    Log(INFO, "Listed documents from collection '%s': %+v", b.CollectionName, results)
    return nextPageToken, nil
}

func (b *BaseModel) Last(ctx context.Context, model interface{}) error {
    if err := b.EnsureCollection(); err != nil {
        Log(ERROR, "Last failed: %v", err)
        return err
    }

    // Query for documents that are not deleted, ordered by creation time descending.
    query := Client.Collection(b.CollectionName).
        Where("deleted", "==", false).
        OrderBy("created_at", firestore.Desc).
        Limit(1)

    iter := query.Documents(ctx)
    defer iter.Stop()

    doc, err := iter.Next()
    if err == iterator.Done {
        Log(WARN, "No records found in collection '%s'", b.CollectionName)
        return errors.New("no records found")
    }
    if err != nil {
        Log(ERROR, "Error fetching last record: %v", err)
        return err
    }

    if err := doc.DataTo(model); err != nil {
        Log(ERROR, "Error mapping document data to model: %v", err)
        return err
    }

    Log(INFO, "Fetched last record from collection '%s': %+v", b.CollectionName, model)
    return nil
}

// Count retrieves the number of documents in the collection that match the provided filters.
func (b *BaseModel) Count(ctx context.Context, filters map[string]interface{}) (int, error) {
	if err := b.EnsureCollection(); err != nil {
		Log(ERROR, "Count failed: %v", err)
		return 0, err
	}

	// Start with a query that excludes deleted documents.
	query := Client.Collection(b.CollectionName).Where("deleted", "==", false)

	// Apply operator filters for range comparisons.
	var err error
	query, err = applyOperatorFilters(query, filters)
	if err != nil {
		Log(ERROR, "Count failed when applying filters: %v", err)
		return 0, err
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			Log(ERROR, "Error iterating documents for count: %v", err)
			return 0, err
		}
		count++
	}

	Log(INFO, "Counted %d documents in collection '%s' with filters: %v", count, b.CollectionName, filters)
	return count, nil
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
	// Leave UpdatedAt as nil on creation.
	b.UpdatedAt = &now
	b.DeletedAt = nil
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

	// Include BaseModel fields explicitly
	baseModelFields := map[string]bool{
		"deleted":    true,
		"deleted_at": true,
		"updated_at": true,
	}

	// Validate updates
	for updateKey, value := range updates {
		if baseModelFields[updateKey] {
			// Allow BaseModel fields
			continue
		}

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

// applyOperatorFilters applies filters that use an operator notation (e.g., "__gt", "__lte").
// If a filter value is a string, it attempts to parse it as a date using the "2006-01-02" layout.
func applyOperatorFilters(query firestore.Query, filters map[string]interface{}) (firestore.Query, error) {
    for key, value := range filters {
        field, op, newValue, err := parseFilter(key, value)
        if err != nil {
            return query, err
        }
        query = query.Where(field, op, newValue)
    }
    return query, nil
}

// parseFilter extracts the field name, operator, and new value from a filter key/value.
// If no operator suffix is provided and the value is a slice, it sets the operator to "in".
func parseFilter(key string, value interface{}) (field string, op string, newValue interface{}, err error) {
    // Simple filter case: no "__" in key.
    if !strings.Contains(key, "__") {
        field = key
        v := reflect.ValueOf(value)
        if v.Kind() == reflect.Slice {
            op = "in"
            newValue = value
            return field, op, newValue, nil
        }
        op = "=="
        // For simple filters, try date parsing but do not return error if it fails.
        if dateStr, ok := value.(string); ok {
            if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
                newValue = parsedDate
            } else {
                newValue = value
            }
        } else {
            newValue = value
        }
        return field, op, newValue, nil
    }

    // Operator filter case: key contains "__".
    parts := strings.SplitN(key, "__", 2)
    field = parts[0]
    switch parts[1] {
    case "gt":
        op = ">"
    case "gte":
        op = ">="
    case "lt":
        op = "<"
    case "lte":
        op = "<="
    default:
        op = "=="
    }
    // For operator filters, if the value is a string, try to parse it as a date.
    if dateStr, ok := value.(string); ok {
        parsedDate, err := time.Parse("2006-01-02", dateStr)
        if err != nil {
            return "", "", nil, fmt.Errorf("invalid date format for %s: %v", key, err)
        }
        newValue = parsedDate
    } else {
        newValue = value
    }
    return field, op, newValue, nil
}
