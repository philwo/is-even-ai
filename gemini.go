// Copyright 2025 Google LLC

// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file or at https://opensource.org/licenses/MIT.

package is_even_ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const geminiSystemPrompt = "You are an AI assistant designed to answer questions about numbers. You will only answer with only the word true or false."

// DefaultGeminiPromptTemplates provides standard prompt templates suitable for Gemini.
var DefaultGeminiPromptTemplates = IsEvenAiCorePromptTemplates{
	IsEven:        func(n int) string { return fmt.Sprintf("Is %d an even number?", n) },
	IsOdd:         func(n int) string { return fmt.Sprintf("Is %d an odd number?", n) },
	AreEqual:      func(a, b int) string { return fmt.Sprintf("Are %d and %d equal?", a, b) },
	AreNotEqual:   func(a, b int) string { return fmt.Sprintf("Are %d and %d not equal?", a, b) },
	IsGreaterThan: func(a, b int) string { return fmt.Sprintf("Is %d greater than %d?", a, b) },
	IsLessThan:    func(a, b int) string { return fmt.Sprintf("Is %d less than %d?", a, b) },
}

// GeminiClientOptions holds configuration for the Gemini client.
type GeminiClientOptions struct {
	APIKey  string
	BaseURL string // Optional: To override the default Gemini API endpoint
}

// GeminiModelOptions specifies options for the Gemini model.
type GeminiModelOptions struct {
	Model       string
	Temperature *float32 // Pointer to allow distinguishing between 0 and not set.
}

// IsEvenAiGemini is an implementation of IsEvenAiCore using the Gemini API.
type IsEvenAiGemini struct {
	*IsEvenAiCore
	genaiModel  *genai.GenerativeModel
	genaiClient *genai.Client
	apiKey      string
	modelName   string
}

// NewIsEvenAiGemini creates a new IsEvenAiGemini client.
func NewIsEvenAiGemini(clientOpts GeminiClientOptions, modelConfigOpts ...GeminiModelOptions) (*IsEvenAiGemini, error) {
	if clientOpts.APIKey == "" {
		return nil, errors.New("gemini API key is required")
	}

	opts := []option.ClientOption{option.WithAPIKey(clientOpts.APIKey)}
	if clientOpts.BaseURL != "" {
		opts = append(opts, option.WithEndpoint(clientOpts.BaseURL))
	}

	// Use a context with timeout for client creation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	createdGenaiClient, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	config := GeminiModelOptions{
		Model: "gemini-2.0-flash-lite", // Default model
	}
	var defaultTemp float32 = 0.0
	config.Temperature = &defaultTemp

	if len(modelConfigOpts) > 0 {
		if modelConfigOpts[0].Model != "" {
			config.Model = modelConfigOpts[0].Model
		}
		if modelConfigOpts[0].Temperature != nil {
			config.Temperature = modelConfigOpts[0].Temperature
		}
	}

	genaiModel := createdGenaiClient.GenerativeModel(config.Model)
	genaiModel.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(geminiSystemPrompt)},
	}

	if config.Temperature != nil {
		genaiModel.SetTemperature(*config.Temperature)
	}

	ai := &IsEvenAiGemini{
		apiKey:      clientOpts.APIKey,
		genaiModel:  genaiModel,
		genaiClient: createdGenaiClient,
		modelName:   config.Model,
	}

	// Each API call gets its own context with a timeout. This makes the query robust
	// against network issues for individual calls and independent of the client creation context.
	queryFunc := func(prompt string) (*bool, error) {
		apiCallCtx, apiCallCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer apiCallCancel()

		resp, err := ai.genaiModel.GenerateContent(apiCallCtx, genai.Text(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate content from Gemini API: %w", err)
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
				return nil, fmt.Errorf("gemini API request blocked, reason: %s", resp.PromptFeedback.BlockReason.String())
			}
			return nil, nil // Undefined response
		}

		part := resp.Candidates[0].Content.Parts[0]
		textContent, ok := part.(genai.Text)
		if !ok {
			return nil, fmt.Errorf("unexpected response part type: %T from Gemini API. Content: %+v", part, resp.Candidates[0].Content.Parts)
		}

		responseContent := strings.ToLower(strings.TrimSpace(string(textContent)))

		switch responseContent {
		case "true":
			b := true
			return &b, nil
		case "false":
			b := false
			return &b, nil
		default:
			return nil, nil
		}
	}

	ai.IsEvenAiCore = NewIsEvenAiCore(DefaultGeminiPromptTemplates, queryFunc)
	return ai, nil
}

// Close client connections if any were long-lived.
func (ai *IsEvenAiGemini) Close() error {
	if ai.genaiClient != nil {
		return ai.genaiClient.Close()
	}
	return nil
}
