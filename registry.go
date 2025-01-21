package firegorm

import (
	"errors"
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

	modelName := modelType.Name()
	if _, exists := modelRegistry[modelName]; exists {
		return errors.New("model already registered: " + modelName)
	}

	modelRegistry[modelName] = ModelInfo{
		CollectionName: collectionName,
		Schema:         modelType,
	}
	return nil
}

// GetModelInfo retrieves the model information for a given model name.
func GetModelInfo(modelName string) (ModelInfo, error) {
	info, exists := modelRegistry[modelName]
	if !exists {
		return ModelInfo{}, errors.New("model not registered: " + modelName)
	}
	return info, nil
}
