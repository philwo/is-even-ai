// Copyright 2025 Google LLC

// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file or at https://opensource.org/licenses/MIT.

package is_even_ai

import (
	"errors"
	"fmt"
	"log" // Added for logging Close errors, if desired
	"sync"
)

var (
	globalGeminiInstance *IsEvenAiGemini // Changed from globalOpenAiInstance
	globalMu             sync.Mutex
	apiKeyIsSet          bool
)

// SetAPIKey configures the global Gemini client instance with the provided API key.
// It must be called before using the convenience functions.
// Additional GeminiModelOptions can be provided to customize model, temperature, etc.
func SetAPIKey(apiKey string, modelOpts ...GeminiModelOptions) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if apiKey == "" {
		apiKeyIsSet = false
		if globalGeminiInstance != nil {
			if err := globalGeminiInstance.Close(); err != nil { // Checked error
				// Optionally log this error, though in many cases, cleanup errors are ignored
				log.Printf("Error closing previous globalGeminiInstance: %v", err)
			}
		}
		globalGeminiInstance = nil
		return errors.New("API key cannot be empty")
	}

	clientOptions := GeminiClientOptions{APIKey: apiKey}

	var mo GeminiModelOptions
	if len(modelOpts) > 0 {
		mo = modelOpts[0]
	}
	// Defaults for model and temperature are set in NewIsEvenAiGemini if not provided here

	instance, err := NewIsEvenAiGemini(clientOptions, mo)
	if err != nil {
		apiKeyIsSet = false
		if globalGeminiInstance != nil {
			if errClose := globalGeminiInstance.Close(); errClose != nil { // Checked error
				log.Printf("Error closing existing globalGeminiInstance on failure: %v", errClose)
			}
		}
		globalGeminiInstance = nil // Ensure instance is nil on error
		return fmt.Errorf("failed to initialize global IsEvenAiGemini instance: %w", err)
	}
	if globalGeminiInstance != nil { // Close previous instance if any
		if errClose := globalGeminiInstance.Close(); errClose != nil { // Checked error
			log.Printf("Error closing previous globalGeminiInstance before new assignment: %v", errClose)
		}
	}
	globalGeminiInstance = instance
	apiKeyIsSet = true
	return nil
}

func getGlobalGeminiInstance() (*IsEvenAiGemini, error) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if !apiKeyIsSet || globalGeminiInstance == nil {
		return nil, errors.New("gemini API key not set or instance not initialized. Call SetAPIKey() first") // Removed period
	}
	return globalGeminiInstance, nil
}

// IsEven checks if n is even using the global Gemini instance.
// Returns *bool (true, false, or nil for undefined) and an error if the operation fails.
func IsEven(n int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsEven(n)
}

// IsOdd checks if n is odd using the global Gemini instance.
func IsOdd(n int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsOdd(n)
}

// AreEqual checks if a and b are equal using the global Gemini instance.
func AreEqual(a, b int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.AreEqual(a, b)
}

// AreNotEqual checks if a and b are not equal using the global Gemini instance.
func AreNotEqual(a, b int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.AreNotEqual(a, b)
}

// IsGreaterThan checks if a is greater than b using the global Gemini instance.
func IsGreaterThan(a, b int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsGreaterThan(a, b)
}

// IsLessThan checks if a is less than b using the global Gemini instance.
func IsLessThan(a, b int) (*bool, error) {
	client, err := getGlobalGeminiInstance()
	if err != nil {
		return nil, err
	}
	return client.IsLessThan(a, b)
}
