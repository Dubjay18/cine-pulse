package model

import "context"

// ModelInterface defines the contract for any AI model
type ModelInterface interface {
	// GenerateText generates text based on the input prompt
	GenerateText(ctx context.Context, prompt string) (string, error)

	// GenerateTextWithOptions generates text with additional options
	GenerateTextWithOptions(ctx context.Context, prompt string, options *GenerationOptions) (string, error)

	// GetModelName returns the name/identifier of the model
	GetModelName() string

	// Close cleans up any resources used by the model
	Close() error
}

// GenerationOptions provides configuration for text generation
type GenerationOptions struct {
	MaxTokens      int      `json:"max_tokens,omitempty"`
	Temperature    float32  `json:"temperature,omitempty"`
	TopP           float32  `json:"top_p,omitempty"`
	TopK           int      `json:"top_k,omitempty"`
	StopSequences  []string `json:"stop_sequences,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"` // "text", "json", etc.
}

// DefaultGenerationOptions returns sensible default options
func DefaultGenerationOptions() *GenerationOptions {
	return &GenerationOptions{
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
	}
}

// ModelConfig holds common configuration for models
type ModelConfig struct {
	APIKey     string
	BaseURL    string
	ModelName  string
	Timeout    int // in seconds
	MaxRetries int
}

// ModelFactory is a factory interface for creating models
type ModelFactory interface {
	CreateModel(config *ModelConfig) (ModelInterface, error)
	GetSupportedModels() []string
}
