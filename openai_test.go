package is_even_ai

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Helper function to check boolean pointer results for OpenAI tests
func checkOpenAIResult(t *testing.T, val *bool, err error, expected bool, funcName string, inputs ...int) {
	t.Helper()
	if err != nil {
		t.Errorf("%s(%v) returned error: %v", funcName, inputs, err)
		return
	}
	if val == nil {
		t.Errorf("%s(%v) returned nil, expected %t", funcName, inputs, expected)
		return
	}
	if *val != expected {
		t.Errorf("%s(%v) = %t; want %t", funcName, inputs, *val, expected)
	}
}

func TestIsEvenAiOpenAi_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration tests: OPENAI_API_KEY not set")
	}

	clientOpts := OpenAIClientOptions{APIKey: apiKey}
	chatOpts := OpenAIChatOptions{Model: "gpt-3.5-turbo", Temperature: 0} // Use deterministic settings

	ai, err := NewIsEvenAiOpenAi(clientOpts, chatOpts)
	if err != nil {
		t.Fatalf("Failed to create NewIsEvenAiOpenAi: %v", err)
	}

	// Test cases mirroring IsEvenAiOpenAi.test.ts
	t.Run("IsEven", func(t *testing.T) {
		res, err := ai.IsEven(2)
		checkOpenAIResult(t, res, err, true, "IsEven", 2)
		res, err = ai.IsEven(3)
		checkOpenAIResult(t, res, err, false, "IsEven", 3)
	})

	t.Run("IsOdd", func(t *testing.T) {
		res, err := ai.IsOdd(4)
		checkOpenAIResult(t, res, err, false, "IsOdd", 4)
		res, err = ai.IsOdd(5)
		checkOpenAIResult(t, res, err, true, "IsOdd", 5)
	})

	t.Run("AreEqual", func(t *testing.T) {
		res, err := ai.AreEqual(6, 6)
		checkOpenAIResult(t, res, err, true, "AreEqual", 6, 6)
		res, err = ai.AreEqual(6, 7)
		checkOpenAIResult(t, res, err, false, "AreEqual", 6, 7)
	})

	t.Run("AreNotEqual", func(t *testing.T) {
		res, err := ai.AreNotEqual(6, 7)
		checkOpenAIResult(t, res, err, true, "AreNotEqual", 6, 7)
		res, err = ai.AreNotEqual(7, 7)
		checkOpenAIResult(t, res, err, false, "AreNotEqual", 7, 7)
	})

	t.Run("IsGreaterThan", func(t *testing.T) {
		res, err := ai.IsGreaterThan(8, 7)
		checkOpenAIResult(t, res, err, true, "IsGreaterThan", 8, 7)
		res, err = ai.IsGreaterThan(7, 8)
		checkOpenAIResult(t, res, err, false, "IsGreaterThan", 7, 8)
	})

	t.Run("IsLessThan", func(t *testing.T) {
		res, err := ai.IsLessThan(8, 9)
		checkOpenAIResult(t, res, err, true, "IsLessThan", 8, 9)
		res, err = ai.IsLessThan(9, 8)
		checkOpenAIResult(t, res, err, false, "IsLessThan", 9, 8)
	})
}

func TestNewIsEvenAiOpenAi_Options(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI options test: OPENAI_API_KEY not set")
	}

	t.Run("DefaultOptions", func(t *testing.T) {
		clientOpts := OpenAIClientOptions{APIKey: apiKey}
		ai, err := NewIsEvenAiOpenAi(clientOpts) // No chat options
		if err != nil {
			t.Fatalf("NewIsEvenAiOpenAi failed: %v", err)
		}
		if ai.chatOptions.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected default model gpt-3.5-turbo, got %s", ai.chatOptions.Model)
		}
		if ai.chatOptions.Temperature != 0 {
			t.Errorf("Expected default temperature 0, got %f", ai.chatOptions.Temperature)
		}
		if ai.httpClient.Timeout != 30*time.Second {
			t.Errorf("Expected default timeout 30s, got %v", ai.httpClient.Timeout)
		}
	})

	t.Run("CustomChatOptions", func(t *testing.T) {
		clientOpts := OpenAIClientOptions{APIKey: apiKey}
		customChatOpts := OpenAIChatOptions{Model: "gpt-4", Temperature: 0.5}
		ai, err := NewIsEvenAiOpenAi(clientOpts, customChatOpts)
		if err != nil {
			t.Fatalf("NewIsEvenAiOpenAi failed: %v", err)
		}
		if ai.chatOptions.Model != "gpt-4" {
			t.Errorf("Expected custom model gpt-4, got %s", ai.chatOptions.Model)
		}
		if ai.chatOptions.Temperature != 0.5 {
			t.Errorf("Expected custom temperature 0.5, got %f", ai.chatOptions.Temperature)
		}
	})

	t.Run("CustomClientOptions", func(t *testing.T) {
		customTimeout := 15 * time.Second
		customBaseURL := "https://api.example.com" // Mock base URL
		clientOpts := OpenAIClientOptions{APIKey: apiKey, Timeout: customTimeout, BaseURL: customBaseURL}
		ai, err := NewIsEvenAiOpenAi(clientOpts)
		if err != nil {
			t.Fatalf("NewIsEvenAiOpenAi failed: %v", err)
		}
		if ai.httpClient.Timeout != customTimeout {
			t.Errorf("Expected custom timeout %v, got %v", customTimeout, ai.httpClient.Timeout)
		}
		expectedEndpoint := "https://api.example.com/v1/chat/completions"
		if ai.openAIEndpoint != expectedEndpoint {
			t.Errorf("Expected custom endpoint %s, got %s", expectedEndpoint, ai.openAIEndpoint)
		}
	})

	t.Run("EmptyAPIKey", func(t *testing.T) {
		_, err := NewIsEvenAiOpenAi(OpenAIClientOptions{APIKey: ""})
		if err == nil {
			t.Error("Expected error for empty API key, got nil")
		} else if err.Error() != "OpenAI API key is required" {
			t.Errorf("Expected error 'OpenAI API key is required', got '%s'", err.Error())
		}
	})
}

// Test to ensure a non-200 response from OpenAI is handled.
// This requires a way to mock the HTTP client or endpoint,
// or have a test API key that predictably fails in a certain way.
// For simplicity, this example doesn't mock http client deeply but one could use http.ServeMux
// and set BaseURL to the test server.
func TestIsEvenAiOpenAi_APIFailure(t *testing.T) {
	apiKey := "test-invalid-key-for-failure" // Use a key known to fail or mock server
	// To properly test this, you'd typically use httptest.NewServer
	// and point BaseURL to it. For this example, we'll assume a key that might cause auth error.

	// Using a dummy base URL that won't resolve, to simulate a network error or non-200 response scenario
	// A more robust test would use httptest.NewServer
	clientOpts := OpenAIClientOptions{APIKey: apiKey, BaseURL: "http://localhost:12345/nonexistent"}
	ai, err := NewIsEvenAiOpenAi(clientOpts)
	if err != nil {
		// This might fail if API key is validated at construction,
		// but our NewIsEvenAiOpenAi only checks for empty.
		// The actual HTTP error will occur during the query.
	}
	if ai == nil { // if constructor itself failed due to bad base URL parsing (not in current code)
		t.Fatalf("Failed to create IsEvenAiOpenAi instance for APIFailure test: %v", err)
	}

	_, err = ai.IsEven(2)
	if err == nil {
		t.Error("Expected an error from IsEven call with failing API/network, got nil")
	} else {
		t.Logf("Got expected error (this is good for this test case): %v", err)
		// Check for specific error content if possible, e.g., connection refused or HTTP status.
		// Contains "failed to send request to OpenAI API" or "OpenAI API request failed with status"
		if !strings.Contains(err.Error(), "failed to send request") && !strings.Contains(err.Error(), "request failed with status") {
			t.Errorf("Error message does not indicate a network or API status error: %s", err.Error())
		}
	}
}
