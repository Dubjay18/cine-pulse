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

// GeminiModel implements the ModelInterface for Google's Gemini API
type GeminiModel struct {
	apiKey    string
	modelName string
	client    *http.Client
}

// Gemini API structures
type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     *float32 `json:"temperature,omitempty"`
	TopP            *float32 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewGeminiModel creates a new Gemini model instance
func NewGeminiModel(config *ModelConfig) (*GeminiModel, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Gemini model")
	}

	modelName := config.ModelName
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 // default 30 seconds
	}

	return &GeminiModel{
		apiKey:    config.APIKey,
		modelName: modelName,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}, nil
}

// GenerateText generates text using Gemini with default options
func (g *GeminiModel) GenerateText(ctx context.Context, prompt string) (string, error) {
	return g.GenerateTextWithOptions(ctx, prompt, DefaultGenerationOptions())
}

// GenerateTextWithOptions generates text using Gemini with custom options
func (g *GeminiModel) GenerateTextWithOptions(ctx context.Context, prompt string, options *GenerationOptions) (string, error) {
	req := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	if options != nil {
		config := &geminiGenerationConfig{}

		if options.Temperature > 0 {
			config.Temperature = &options.Temperature
		}
		if options.TopP > 0 {
			config.TopP = &options.TopP
		}
		if options.TopK > 0 {
			config.TopK = &options.TopK
		}
		if options.MaxTokens > 0 {
			config.MaxOutputTokens = &options.MaxTokens
		}
		if len(options.StopSequences) > 0 {
			config.StopSequences = options.StopSequences
		}

		req.GenerationConfig = config
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.modelName, g.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no parts in candidate content")
	}

	return candidate.Content.Parts[0].Text, nil
}

// GetModelName returns the name of the Gemini model
func (g *GeminiModel) GetModelName() string {
	return fmt.Sprintf("gemini:%s", g.modelName)
}

// Close cleans up resources (no-op for Gemini HTTP client)
func (g *GeminiModel) Close() error {
	return nil
}

// GeminiFactory implements ModelFactory for Gemini models
type GeminiFactory struct{}

// CreateModel creates a new Gemini model instance
func (f *GeminiFactory) CreateModel(config *ModelConfig) (ModelInterface, error) {
	return NewGeminiModel(config)
}

// GetSupportedModels returns the list of supported Gemini models
func (f *GeminiFactory) GetSupportedModels() []string {
	return []string{
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-1.0-pro",
	}
}

// NewGeminiFactory creates a new Gemini factory
func NewGeminiFactory() ModelFactory {
	return &GeminiFactory{}
}
