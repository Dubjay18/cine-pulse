package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIModel implements the ModelInterface for OpenAI API
type OpenAIModel struct {
	apiKey    string
	baseURL   string
	modelName string
	client    *http.Client
}

// OpenAI API request/response structures
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	TopP        float32   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []choice  `json:"choices"`
	Error   *apiError `json:"error,omitempty"`
}

type choice struct {
	Message message `json:"message"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewOpenAIModel creates a new OpenAI model instance
func NewOpenAIModel(config *ModelConfig) (*OpenAIModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI model")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	modelName := config.ModelName
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 // default 30 seconds
	}

	return &OpenAIModel{
		apiKey:    config.APIKey,
		baseURL:   baseURL,
		modelName: modelName,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}, nil
}

// GenerateText generates text using OpenAI with default options
func (o *OpenAIModel) GenerateText(ctx context.Context, prompt string) (string, error) {
	return o.GenerateTextWithOptions(ctx, prompt, DefaultGenerationOptions())
}

// GenerateTextWithOptions generates text using OpenAI with custom options
func (o *OpenAIModel) GenerateTextWithOptions(ctx context.Context, prompt string, options *GenerationOptions) (string, error) {
	req := openAIRequest{
		Model: o.modelName,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	if options != nil {
		if options.MaxTokens > 0 {
			req.MaxTokens = options.MaxTokens
		}
		if options.Temperature > 0 {
			req.Temperature = options.Temperature
		}
		if options.TopP > 0 {
			req.TopP = options.TopP
		}
		if len(options.StopSequences) > 0 {
			req.Stop = options.StopSequences
		}
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// GetModelName returns the name of the OpenAI model
func (o *OpenAIModel) GetModelName() string {
	return fmt.Sprintf("openai:%s", o.modelName)
}

// Close cleans up resources (no-op for OpenAI)
func (o *OpenAIModel) Close() error {
	return nil
}

// OpenAIFactory implements ModelFactory for OpenAI models
type OpenAIFactory struct{}

// CreateModel creates a new OpenAI model instance
func (f *OpenAIFactory) CreateModel(config *ModelConfig) (ModelInterface, error) {
	return NewOpenAIModel(config)
}

// GetSupportedModels returns the list of supported OpenAI models
func (f *OpenAIFactory) GetSupportedModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// NewOpenAIFactory creates a new OpenAI factory
func NewOpenAIFactory() ModelFactory {
	return &OpenAIFactory{}
}
