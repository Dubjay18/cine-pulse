package model

import (
	"context"
	"fmt"
	"strings"
)

// ModelType represents different types of AI models
type ModelType string

const (
	ModelTypeOpenAI ModelType = "openai"
	ModelTypeGemini ModelType = "gemini"
	// Add more model types as needed
	// ModelTypeClaude  ModelType = "claude"
	// ModelTypeLlama   ModelType = "llama"
)

// ModelManager manages different AI models and provides a unified interface
type ModelManager struct {
	factories map[ModelType]ModelFactory
	models    map[string]ModelInterface
}

// NewModelManager creates a new model manager
func NewModelManager() *ModelManager {
	manager := &ModelManager{
		factories: make(map[ModelType]ModelFactory),
		models:    make(map[string]ModelInterface),
	}

	// Register available factories
	manager.RegisterFactory(ModelTypeOpenAI, NewOpenAIFactory())
	manager.RegisterFactory(ModelTypeGemini, NewGeminiFactory())

	return manager
}

// RegisterFactory registers a model factory
func (m *ModelManager) RegisterFactory(modelType ModelType, factory ModelFactory) {
	m.factories[modelType] = factory
}

// CreateModel creates a model instance
func (m *ModelManager) CreateModel(modelType ModelType, config *ModelConfig) (ModelInterface, error) {
	factory, exists := m.factories[modelType]
	if !exists {
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}

	model, err := factory.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Store the model instance with a unique key
	key := fmt.Sprintf("%s:%s", modelType, config.ModelName)
	m.models[key] = model

	return model, nil
}

// GetModel retrieves a previously created model
func (m *ModelManager) GetModel(modelType ModelType, modelName string) (ModelInterface, bool) {
	key := fmt.Sprintf("%s:%s", modelType, modelName)
	model, exists := m.models[key]
	return model, exists
}

// GetOrCreateModel gets an existing model or creates a new one
func (m *ModelManager) GetOrCreateModel(modelType ModelType, config *ModelConfig) (ModelInterface, error) {
	if model, exists := m.GetModel(modelType, config.ModelName); exists {
		return model, nil
	}
	return m.CreateModel(modelType, config)
}

// ListSupportedModels returns all supported models across all factories
func (m *ModelManager) ListSupportedModels() map[ModelType][]string {
	result := make(map[ModelType][]string)
	for modelType, factory := range m.factories {
		result[modelType] = factory.GetSupportedModels()
	}
	return result
}

// CloseAll closes all model instances
func (m *ModelManager) CloseAll() error {
	var errors []string
	for key, model := range m.models {
		if err := model.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing models: %s", strings.Join(errors, ", "))
	}

	// Clear the models map
	m.models = make(map[string]ModelInterface)
	return nil
}

// GenerateWithBestModel attempts to generate text using multiple models as fallback
func (m *ModelManager) GenerateWithBestModel(ctx context.Context, prompt string, configs []*ModelConfig, options *GenerationOptions) (string, string, error) {
	var lastErr error

	for _, config := range configs {
		// Determine model type from config or model name
		modelType := m.detectModelType(config.ModelName)
		if modelType == "" {
			continue
		}

		model, err := m.GetOrCreateModel(modelType, config)
		if err != nil {
			lastErr = err
			continue
		}

		result, err := model.GenerateTextWithOptions(ctx, prompt, options)
		if err != nil {
			lastErr = err
			continue
		}

		return result, model.GetModelName(), nil
	}

	if lastErr != nil {
		return "", "", fmt.Errorf("all models failed, last error: %w", lastErr)
	}

	return "", "", fmt.Errorf("no valid model configurations provided")
}

// detectModelType attempts to detect the model type from the model name
func (m *ModelManager) detectModelType(modelName string) ModelType {
	modelName = strings.ToLower(modelName)

	if strings.Contains(modelName, "gpt") || strings.Contains(modelName, "openai") {
		return ModelTypeOpenAI
	}
	if strings.Contains(modelName, "gemini") {
		return ModelTypeGemini
	}

	return ""
}
