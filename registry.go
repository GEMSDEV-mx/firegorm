package firegorm

import (
	"fmt"
	"reflect"
	"strings"
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
		Log(WARN, "Model '%s' is already registered", modelName)
		return nil, fmt.Errorf("model '%s' is already registered", modelName)
	}

	// Build tag-to-field mapping
    tagToFieldMap := make(map[string]string)
    for i := 0; i < modelType.NumField(); i++ {
        field := modelType.Field(i)

        // strip options from firestore tag
        rawFS := field.Tag.Get("firestore")
        if rawFS != "" && rawFS != "-" {
            name := strings.SplitN(rawFS, ",", 2)[0]
            if name != "" {
                tagToFieldMap[name] = field.Name
            }
        }

        // strip options from json tag
        rawJSON := field.Tag.Get("json")
        if rawJSON != "" && rawJSON != "-" {
            name := strings.SplitN(rawJSON, ",", 2)[0]
            if name != "" {
                tagToFieldMap[name] = field.Name
            }
        }
    }

	modelRegistry[modelName] = ModelInfo{
		CollectionName: collectionName,
		Schema:         modelType,
		TagToFieldMap:  tagToFieldMap,
	}
	Log(INFO, "Registered model '%s' with collection '%s': %+v", modelName, collectionName, modelRegistry)

	// Initialize the model instance
	instance := reflect.New(modelType).Interface()

	// Set collection name and model name
	if baseModel, ok := instance.(interface {
		SetCollectionName(string)
		SetModelName(string)
	}); ok {
		baseModel.SetCollectionName(collectionName)
		baseModel.SetModelName(modelType.Name())
		Log(DEBUG, "Initialized model instance with collection '%s' and name '%s'", collectionName, modelType.Name())
	}

	return instance, nil
}

// GetModelInfo retrieves metadata for a registered model.
func GetModelInfo(modelName string) (ModelInfo, error) {
	info, exists := modelRegistry[modelName]
	if !exists {
		Log(WARN, "Model '%s' is not registered. Current registry: %+v", modelName, modelRegistry)
		return ModelInfo{}, fmt.Errorf("model '%s' is not registered", modelName)
	}
	Log(DEBUG, "Retrieved model info for '%s': %+v", modelName, info)
	return info, nil
}
