package firegorm

import (
	"fmt"
	"log"
	"reflect"
)

// ModelInfo stores metadata for registered models.
type ModelInfo struct {
	CollectionName string
	Schema         reflect.Type
	TagToFieldMap  map[string]string // Maps Firestore/JSON tags to field names
}

// Registry to store models and their metadata.
var modelRegistry = make(map[string]ModelInfo)

// RegisterModel registers a model with its collection name and schema.
func RegisterModel(model interface{}, collectionName string) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Use collectionName as the key
	if _, exists := modelRegistry[collectionName]; exists {
		return fmt.Errorf("collection '%s' is already registered", collectionName)
	}

	// Build tag-to-field mapping
	tagToFieldMap := make(map[string]string)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		firestoreTag := field.Tag.Get("firestore")
		jsonTag := field.Tag.Get("json")

		if firestoreTag != "" && firestoreTag != "-" {
			tagToFieldMap[firestoreTag] = field.Name
		}
		if jsonTag != "" && jsonTag != "-" {
			tagToFieldMap[jsonTag] = field.Name
		}
	}

	// Store the model in the registry
	modelRegistry[collectionName] = ModelInfo{
		CollectionName: collectionName,
		Schema:         modelType,
		TagToFieldMap:  tagToFieldMap,
	}
	log.Printf("Registered model for collection '%s': %+v", collectionName, modelRegistry[collectionName])
	return nil
}


// GetModelInfo retrieves metadata for a registered model by its collection name.
func GetModelInfo(collectionName string) (ModelInfo, error) {
	info, exists := modelRegistry[collectionName]
	if !exists {
		log.Printf("Collection '%s' is not registered. Current registry: %+v", collectionName, modelRegistry)
		return ModelInfo{}, fmt.Errorf("collection '%s' is not registered", collectionName)
	}
	return info, nil
}
