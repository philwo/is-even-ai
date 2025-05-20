package is_even_ai

import (
	"errors"
	"fmt"
	"sync"
)

var (
	globalOpenAiInstance *IsEvenAiOpenAi
	globalMu             sync.Mutex
	apiKeyIsSet          bool
)

// SetAPIKey configures the global OpenAI client instance with the provided API key.
// It must be called before using the convenience functions.
// Additional OpenAIChatOptions can be provided to customize model, temperature, etc.
func SetAPIKey(apiKey string, chatOpts ...OpenAIChatOptions) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if apiKey == "" {
		apiKeyIsSet = false
		// Clear instance if API key is unset, or handle as error
		globalOpenAiInstance = nil
		return errors.New("API key cannot be empty")
	}

	clientOptions := OpenAIClientOptions{APIKey: apiKey}

	var co OpenAIChatOptions
	if len(chatOpts) > 0 {
		co = chatOpts[0]
	}
	// Defaults for model and temperature are set in NewIsEvenAiOpenAi if not provided here

	instance, err := NewIsEvenAiOpenAi(clientOptions, co)
	if err != nil {
		apiKeyIsSet = false
		globalOpenAiInstance = nil // Ensure instance is nil on error
		return fmt.Errorf("failed to initialize global IsEvenAiOpenAi instance: %w", err)
	}
	globalOpenAiInstance = instance
	apiKeyIsSet = true
	return nil
}

func getGlobalOpenAiInstance() (*IsEvenAiOpenAi, error) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if !apiKeyIsSet || globalOpenAiInstance == nil {
		return nil, errors.New("OpenAI API key not set or instance not initialized. Call SetAPIKey() first.")
	}
	return globalOpenAiInstance, nil
}

// IsEven checks if n is even using the global OpenAI instance.
// Returns *bool (true, false, or nil for undefined) and an error if the operation fails.
func IsEven(n int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsEven(n)
}

// IsOdd checks if n is odd using the global OpenAI instance.
func IsOdd(n int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsOdd(n)
}

// AreEqual checks if a and b are equal using the global OpenAI instance.
func AreEqual(a, b int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.AreEqual(a, b)
}

// AreNotEqual checks if a and b are not equal using the global OpenAI instance.
func AreNotEqual(a, b int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.AreNotEqual(a, b)
}

// IsGreaterThan checks if a is greater than b using the global OpenAI instance.
func IsGreaterThan(a, b int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsGreaterThan(a, b)
}

// IsLessThan checks if a is less than b using the global OpenAI instance.
func IsLessThan(a, b int) (*bool, error) {
	client, err := getGlobalOpenAiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsLessThan(a, b)
}
