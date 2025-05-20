package is_even_ai

import (
	"os"
	"testing"
)

// Helper to reset global state for convenience tests
func resetGlobalState() {
	globalMu.Lock()
	globalOpenAiInstance = nil
	apiKeyIsSet = false
	globalMu.Unlock()
}

// Helper function to check boolean pointer results
func checkConvenienceResult(t *testing.T, val *bool, err error, expected bool, funcName string, inputs ...int) {
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

func TestConvenience_SetAPIKeyAndUse(t *testing.T) {
	originalApiKey := os.Getenv("OPENAI_API_KEY")
	apiKeyForTest := "test-api-key-from-setapikey" // Use a distinct key for this test

	if originalApiKey == "" {
		t.Log("OPENAI_API_KEY not in env, SetAPIKey test will use a dummy key for instantiation logic only, API calls will likely fail if not mocked.")
		// If we want this test to pass without a real key for basic SetAPIKey logic,
		// we might need a mock HTTP transport or expect errors from the functions.
		// For now, we proceed assuming SetAPIKey should succeed in creating an instance.
		// The actual calls to IsEven, etc., will fail if this dummy key is invalid.
		// Let's use a known valid key if available for a fuller test.
		// If a real OPENAI_API_KEY is needed for these to pass, this test configuration needs adjustment
		// or the test scope reduced to just SetAPIKey's effect on global vars.
		// Given convenience.test.ts expects actual results, a valid key is implicitly needed.
		// For robust testing without hitting API, a mock HTTP client at global level would be needed.
		// This test will use a real key if available.
		if os.Getenv("OPENAI_API_KEY_FOR_TESTS") != "" {
			apiKeyForTest = os.Getenv("OPENAI_API_KEY_FOR_TESTS")
		} else {
			t.Skip("Skipping TestConvenience_SetAPIKeyAndUse: OPENAI_API_KEY_FOR_TESTS not set, and no fallback for full convenience function test without a real key.")
			return
		}
	} else {
		// If OPENAI_API_KEY is set, use that to ensure functions actually work.
		apiKeyForTest = originalApiKey
	}

	// Test 1: Set API Key using SetAPIKey function
	t.Run("WithKeyPassedToSetAPIKey", func(t *testing.T) {
		resetGlobalState()
		// Temporarily unset OPENAI_API_KEY from environment if it was there
		// to ensure SetAPIKey is the one providing the key.
		if originalApiKey != "" {
			currentEnvKey := os.Getenv("OPENAI_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")
			defer os.Setenv("OPENAI_API_KEY", currentEnvKey) // Restore
		}

		err := SetAPIKey(apiKeyForTest)
		if err != nil {
			t.Fatalf("SetAPIKey failed: %v", err)
		}
		if !apiKeyIsSet {
			t.Fatal("apiKeyIsSet should be true after SetAPIKey")
		}
		if globalOpenAiInstance == nil {
			t.Fatal("globalOpenAiInstance should be initialized after SetAPIKey")
		}
		if globalOpenAiInstance.apiKey != apiKeyForTest {
			t.Fatalf("globalOpenAiInstance.apiKey = %s; want %s", globalOpenAiInstance.apiKey, apiKeyForTest)
		}

		// Test convenience functions
		resBool, errBool := IsEven(2)
		checkConvenienceResult(t, resBool, errBool, true, "IsEven", 2)
		resBool, errBool = IsEven(3)
		checkConvenienceResult(t, resBool, errBool, false, "IsEven", 3)

		resBool, errBool = IsOdd(4)
		checkConvenienceResult(t, resBool, errBool, false, "IsOdd", 4)
		resBool, errBool = IsOdd(5)
		checkConvenienceResult(t, resBool, errBool, true, "IsOdd", 5)

		resBool, errBool = AreEqual(6, 6)
		checkConvenienceResult(t, resBool, errBool, true, "AreEqual", 6, 6)
		resBool, errBool = AreEqual(6, 7)
		checkConvenienceResult(t, resBool, errBool, false, "AreEqual", 6, 7)

		resBool, errBool = AreNotEqual(6, 7)
		checkConvenienceResult(t, resBool, errBool, true, "AreNotEqual", 6, 7)
		resBool, errBool = AreNotEqual(7, 7)
		checkConvenienceResult(t, resBool, errBool, false, "AreNotEqual", 7, 7)

		resBool, errBool = IsGreaterThan(8, 7)
		checkConvenienceResult(t, resBool, errBool, true, "IsGreaterThan", 8, 7)
		resBool, errBool = IsGreaterThan(7, 8)
		checkConvenienceResult(t, resBool, errBool, false, "IsGreaterThan", 7, 8)

		resBool, errBool = IsLessThan(8, 9)
		checkConvenienceResult(t, resBool, errBool, true, "IsLessThan", 8, 9)
		resBool, errBool = IsLessThan(9, 8)
		checkConvenienceResult(t, resBool, errBool, false, "IsLessThan", 9, 8)

		resetGlobalState() // Clean up for next test
	})
}

func TestConvenience_ApiKeyFromEnv(t *testing.T) {
	resetGlobalState()
	originalApiKey := os.Getenv("OPENAI_API_KEY")

	if originalApiKey == "" {
		t.Skip("Skipping TestConvenience_ApiKeyFromEnv: OPENAI_API_KEY not set in environment.")
		return
	}
	// For this test, SetAPIKey should pick up the env var if called with it.
	// The convenience functions themselves rely on SetAPIKey being called.
	// The example main.go shows os.Getenv and then calls SetAPIKey.
	// So, this test will also call SetAPIKey with the env-retrieved key.

	err := SetAPIKey(originalApiKey)
	if err != nil {
		t.Fatalf("SetAPIKey with env key failed: %v", err)
	}

	// Test convenience functions (sample)
	resBool, errBool := IsEven(20)
	checkConvenienceResult(t, resBool, errBool, true, "IsEven", 20)
	resBool, errBool = IsOdd(21)
	checkConvenienceResult(t, resBool, errBool, true, "IsOdd", 21)

	resetGlobalState()
}

func TestConvenience_NoAPIKeySet(t *testing.T) {
	resetGlobalState() // Ensure no key is set

	// Attempt to use a convenience function without setting API key
	_, err := IsEven(2)
	if err == nil {
		t.Fatal("Expected error when calling IsEven without API key, got nil")
	}
	expectedErrorMsg := "OpenAI API key not set or instance not initialized. Call SetAPIKey() first."
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	// Test SetAPIKey with empty string
	err = SetAPIKey("")
	if err == nil {
		t.Fatal("Expected error when calling SetAPIKey with empty string, got nil")
	}
	if err.Error() != "API key cannot be empty" {
		t.Errorf("Expected error 'API key cannot be empty', got '%s'", err.Error())
	}
	if apiKeyIsSet {
		t.Error("apiKeyIsSet should be false after SetAPIKey with empty string")
	}
	if globalOpenAiInstance != nil {
		t.Error("globalOpenAiInstance should be nil after SetAPIKey with empty string")
	}
}

func TestConvenience_SetAPIKeyWithChatOptions(t *testing.T) {
	resetGlobalState()
	apiKey := "test-key-for-options"
	if os.Getenv("OPENAI_API_KEY_FOR_TESTS") != "" {
		apiKey = os.Getenv("OPENAI_API_KEY_FOR_TESTS")
	} else if os.Getenv("OPENAI_API_KEY") != "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	} else {
		t.Skip("Skipping TestConvenience_SetAPIKeyWithChatOptions: No API key available for testing instantiation with options.")
		return
	}

	customOpts := OpenAIChatOptions{Model: "gpt-4-turbo", Temperature: 0.7}
	err := SetAPIKey(apiKey, customOpts)
	if err != nil {
		t.Fatalf("SetAPIKey with custom chat options failed: %v", err)
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	if globalOpenAiInstance == nil {
		t.Fatal("globalOpenAiInstance is nil after SetAPIKey with custom options")
	}
	if globalOpenAiInstance.chatOptions.Model != customOpts.Model {
		t.Errorf("Expected model %s, got %s", customOpts.Model, globalOpenAiInstance.chatOptions.Model)
	}
	if globalOpenAiInstance.chatOptions.Temperature != customOpts.Temperature {
		t.Errorf("Expected temperature %f, got %f", customOpts.Temperature, globalOpenAiInstance.chatOptions.Temperature)
	}
	resetGlobalState()
}
