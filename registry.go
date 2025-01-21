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
func RegisterModel(model interface{}, collectionName string) (interface{}, error) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	modelName := collectionName + "." + modelType.Name()

	if _, exists := modelRegistry[modelName]; exists {
		return nil,fmt.Errorf("model '%s' is already registered", modelName)
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

	modelRegistry[modelName] = ModelInfo{
		CollectionName: collectionName,
		Schema:         modelType,
		TagToFieldMap:  tagToFieldMap,
	}
	log.Printf("Registered model '%s' with collection '%s': %+v", modelName, collectionName, modelRegistry)

	// Initialize the model instance
	instance := reflect.New(modelType).Interface()

	// Set collection name and model name
	if baseModel, ok := instance.(interface {
		SetCollectionName(string)
		SetModelName(string)
	}); ok {
		baseModel.SetCollectionName(collectionName)
		baseModel.SetModelName(modelType.Name())
	}

	return instance, nil
}

// GetModelInfo retrieves metadata for a registered model.
func GetModelInfo(modelName string) (ModelInfo, error) {
	info, exists := modelRegistry[modelName]
	if !exists {
		log.Printf("Model '%s' is not registered. Current registry: %+v", modelName, modelRegistry)
		return ModelInfo{}, fmt.Errorf("model '%s' is not registered", modelName)
	}
	return info, nil
}
