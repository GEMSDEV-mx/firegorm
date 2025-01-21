package firegorm

import (
	"errors"
	"log"
	"reflect"
)

// ModelInfo stores the schema and collection name for a registered model.
type ModelInfo struct {
	CollectionName string
	Schema         reflect.Type
}

// modelRegistry stores all registered models.
var modelRegistry = make(map[string]ModelInfo)

// RegisterModel registers a model with its collection name and schema.
func RegisterModel(model interface{}, collectionName string) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Use the fully qualified name (package + type) as the model key
	modelName := modelType.PkgPath() + "." + modelType.Name()
	if _, exists := modelRegistry[modelName]; exists {
		log.Printf("Model '%s' is already registered for collection '%s'.", modelName, modelRegistry[modelName].CollectionName)
		return errors.New("model already registered: " + modelName)
	}

	modelRegistry[modelName] = ModelInfo{
		CollectionName: collectionName,
		Schema:         modelType,
	}

	log.Printf("Successfully registered model '%s' with collection '%s'.", modelName, collectionName)
	return nil
}

// GetModelInfo retrieves the model information for a given model name.
func GetModelInfo(modelName string) (ModelInfo, error) {
	info, exists := modelRegistry[modelName]
	if !exists {
		log.Printf("Model '%s' is not registered.", modelName)
		return ModelInfo{}, errors.New("model not registered: " + modelName)
	}
	return info, nil
}
