package is_even_ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	genaiModel *genai.GenerativeModel
	apiKey     string
}

// NewIsEvenAiGemini creates a new IsEvenAiGemini client.
// 'clientOpts' are options for the API key and optional custom endpoint.
// 'modelConfigOpts' can optionally override default model and temperature settings.
func NewIsEvenAiGemini(clientOpts GeminiClientOptions, modelConfigOpts ...GeminiModelOptions) (*IsEvenAiGemini, error) {
	if clientOpts.APIKey == "" {
		return nil, errors.New("Gemini API key is required")
	}

	opts := []option.ClientOption{option.WithAPIKey(clientOpts.APIKey)}
	if clientOpts.BaseURL != "" {
		opts = append(opts, option.WithEndpoint(clientOpts.BaseURL))
	}

	ctx := context.Background()
	genaiClient, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	// Defer closing the client, though for a long-lived IsEvenAiGemini instance,
	// the client should also be long-lived.
	// Consider if the client should be closed when IsEvenAiGemini is no longer needed.
	// For this library structure, we'll assume client is managed with the lifetime of IsEvenAiGemini.
	// To properly manage, IsEvenAiGemini would need a Close() method:
	// defer genaiClient.Close() // This would close it immediately, not intended here.

	config := GeminiModelOptions{
		Model: "gemini-2.0-flash-lite", // Default model
	}
	var defaultTemp float32 = 0.0 // Default temperature for deterministic responses
	config.Temperature = &defaultTemp

	if len(modelConfigOpts) > 0 {
		if modelConfigOpts[0].Model != "" {
			config.Model = modelConfigOpts[0].Model
		}
		if modelConfigOpts[0].Temperature != nil {
			config.Temperature = modelConfigOpts[0].Temperature
		}
	}

	genaiModel := genaiClient.GenerativeModel(config.Model)
	genaiModel.SetSystemInstruction(genai.Text(geminiSystemPrompt))
	if config.Temperature != nil {
		genaiModel.SetTemperature(*config.Temperature)
	}

	ai := &IsEvenAiGemini{
		apiKey:     clientOpts.APIKey,
		genaiModel: genaiModel,
	}

	queryFunc := func(prompt string) (*bool, error) {
		resp, err := ai.genaiModel.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate content from Gemini API: %w", err)
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			// Try to get error details if any
			if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
				return nil, fmt.Errorf("Gemini API request blocked, reason: %s", resp.PromptFeedback.BlockReason.String())
			}
			return nil, nil // Undetermined or empty response
		}

		part := resp.Candidates[0].Content.Parts[0]
		textContent, ok := part.(genai.Text)
		if !ok {
			// Check if it's a function call or other part type, though not expected here.
			return nil, fmt.Errorf("unexpected response part type: %T, content: %+v", part, resp.Candidates[0].Content.Parts)
		}

		responseContent := strings.ToLower(strings.TrimSpace(string(textContent)))

		if responseContent == "true" {
			b := true
			return &b, nil
		} else if responseContent == "false" {
			b := false
			return &b, nil
		}
		// The AI might respond with more than just "true" or "false".
		// e.g. "true." or "The answer is true."
		// For robustness, one might check if the response *contains* "true" or "false"
		// but the system prompt is strict.
		// If the response isn't exactly "true" or "false", treat as undefined.
		return nil, nil // Response was not strictly "true" or "false"
	}

	ai.IsEvenAiCore = NewIsEvenAiCore(DefaultGeminiPromptTemplates, queryFunc)
	return ai, nil
}

// Close client connections if any were long-lived.
// For go-genai, the client has a Close() method.
func (ai *IsEvenAiGemini) Close() error {
	if ai.genaiModel != nil && ai.genaiModel.Client != nil {
		return ai.genaiModel.Client.Close()
	}
	return nil
}
