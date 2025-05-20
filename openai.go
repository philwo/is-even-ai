package is_even_ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const systemPrompt = "You are an AI assistant designed to answer questions about numbers. You will only answer with only the word true or false."

// DefaultOpenAiPromptTemplates provides standard prompt templates suitable for OpenAI.
var DefaultOpenAiPromptTemplates = IsEvenAiCorePromptTemplates{
	IsEven:        func(n int) string { return fmt.Sprintf("Is %d an even number?", n) },
	IsOdd:         func(n int) string { return fmt.Sprintf("Is %d an odd number?", n) },
	AreEqual:      func(a, b int) string { return fmt.Sprintf("Are %d and %d equal?", a, b) },
	AreNotEqual:   func(a, b int) string { return fmt.Sprintf("Are %d and %d not equal?", a, b) },
	IsGreaterThan: func(a, b int) string { return fmt.Sprintf("Is %d greater than %d?", a, b) },
	IsLessThan:    func(a, b int) string { return fmt.Sprintf("Is %d less than %d?", a, b) },
}

// OpenAIClientOptions holds configuration for the OpenAI client.
type OpenAIClientOptions struct {
	APIKey  string
	BaseURL string        // Optional: To override the default OpenAI API base URL
	Timeout time.Duration // Optional: HTTP client timeout
}

// OpenAIChatOptions specifies options for the OpenAI Chat Completion API.
type OpenAIChatOptions struct {
	Model       string
	Temperature float32
	// Other OpenAI parameters like MaxTokens, TopP, etc., can be added here.
}

// IsEvenAiOpenAi is an implementation of IsEvenAiCore using the OpenAI API.
type IsEvenAiOpenAi struct {
	*IsEvenAiCore
	httpClient     *http.Client
	apiKey         string
	chatOptions    OpenAIChatOptions
	openAIEndpoint string
}

// NewIsEvenAiOpenAi creates a new IsEvenAiOpenAi client.
// 'clientOpts' are options for the HTTP client and API key.
// 'chatCompletionOpts' can optionally override default model and temperature settings.
func NewIsEvenAiOpenAi(clientOpts OpenAIClientOptions, chatCompletionOpts ...OpenAIChatOptions) (*IsEvenAiOpenAi, error) {
	if clientOpts.APIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	timeout := clientOpts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	chatOpts := OpenAIChatOptions{
		Model:       "gpt-3.5-turbo", // Default model
		Temperature: 0,               // Default temperature for deterministic responses
	}
	if len(chatCompletionOpts) > 0 {
		chatOpts = chatCompletionOpts[0]
		if chatOpts.Model == "" {
			chatOpts.Model = "gpt-3.5-turbo" // Ensure model is not empty
		}
	}

	apiEndpoint := "https://api.openai.com/v1/chat/completions"
	if clientOpts.BaseURL != "" {
		// Ensure BaseURL ends with a slash if it's going to be joined with /v1/chat/completions
		// or ensure the full path is provided. For simplicity, assuming BaseURL is just the host.
		// A more robust solution would involve url.Parse and url.JoinPath.
		apiEndpoint = strings.TrimRight(clientOpts.BaseURL, "/") + "/v1/chat/completions"
	}

	ai := &IsEvenAiOpenAi{
		httpClient:     httpClient,
		apiKey:         clientOpts.APIKey,
		chatOptions:    chatOpts,
		openAIEndpoint: apiEndpoint,
	}

	// Define the query function that calls the OpenAI API
	queryFunc := func(prompt string) (*bool, error) {
		requestPayload := map[string]interface{}{
			"model":       ai.chatOptions.Model,
			"temperature": ai.chatOptions.Temperature,
			"messages": []map[string]string{
				{"role": "system", "content": systemPrompt},
				{"role": "user", "content": prompt},
			},
			// "stream": true, // For streaming responses, would require different handling
		}
		payloadBytes, err := json.Marshal(requestPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal OpenAI request payload: %w", err)
		}

		req, err := http.NewRequestWithContext(context.Background(), "POST", ai.openAIEndpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ai.apiKey)

		resp, err := ai.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request to OpenAI API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("OpenAI API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// This part handles a non-streaming response.
		// The original TypeScript code uses streaming and checks prefixes.
		// A full Go streaming implementation would be more complex.
		var openAiResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&openAiResp); err != nil {
			return nil, fmt.Errorf("failed to decode OpenAI API response: %w", err)
		}

		if len(openAiResp.Choices) == 0 || openAiResp.Choices[0].Message.Content == "" {
			return nil, nil // Undetermined or empty response
		}

		responseContent := strings.ToLower(strings.TrimSpace(openAiResp.Choices[0].Message.Content))

		// The TypeScript code's streaming logic allows early exit if "true" or "false" is detected.
		// e.g., if ("true".startsWith(response)) return true;
		// Here, we check the full content.
		if responseContent == "true" {
			b := true
			return &b, nil
		} else if responseContent == "false" {
			b := false
			return &b, nil
		}

		return nil, nil // Response was not "true" or "false"
	}

	// Initialize the embedded IsEvenAiCore with the OpenAI-specific query function and default templates
	ai.IsEvenAiCore = NewIsEvenAiCore(DefaultOpenAiPromptTemplates, queryFunc)
	return ai, nil
}
