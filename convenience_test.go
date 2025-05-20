package is_even_ai

import (
	"os"
	"testing"
)

// Helper to reset global state for convenience tests
func resetGlobalStateAndClose() {
	globalMu.Lock()
	if globalGeminiInstance != nil {
		globalGeminiInstance.Close()
		globalGeminiInstance = nil
	}
	apiKeyIsSet = false
	globalMu.Unlock()
}

// Helper function to check boolean pointer results (re-declared or use one from gemini_test.go if visible)
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

func TestConvenience_SetAPIKeyAndUse_Gemini(t *testing.T) {
	originalApiKey := os.Getenv("GEMINI_API_KEY") // Changed to GEMINI_API_KEY
	apiKeyForTest := "test-api-key-from-setapikey-gemini"

	if originalApiKey == "" {
		t.Log("GEMINI_API_KEY not in env, SetAPIKey test will use a dummy key for instantiation logic only.")
		// For robust testing without hitting API, a mock HTTP transport at global level would be needed.
		// This test will use a real key if available.
		if os.Getenv("GEMINI_API_KEY_FOR_TESTS") != "" { // Changed to GEMINI_API_KEY_FOR_TESTS
			apiKeyForTest = os.Getenv("GEMINI_API_KEY_FOR_TESTS")
		} else {
			t.Skip("Skipping TestConvenience_SetAPIKeyAndUse_Gemini: GEMINI_API_KEY_FOR_TESTS not set, and no fallback for full convenience function test without a real key.")
			return
		}
	} else {
		apiKeyForTest = originalApiKey
	}

	t.Run("WithKeyPassedToSetAPIKey_Gemini", func(t *testing.T) {
		resetGlobalStateAndClose()
		// Temporarily unset GEMINI_API_KEY from environment if it was there
		// to ensure SetAPIKey is the one providing the key.
		if originalApiKey != "" {
			currentEnvKey := os.Getenv("GEMINI_API_KEY")
			os.Unsetenv("GEMINI_API_KEY")
			defer os.Setenv("GEMINI_API_KEY", currentEnvKey) // Restore
		}

		err := SetAPIKey(apiKeyForTest)
		if err != nil {
			t.Fatalf("SetAPIKey failed: %v", err)
		}
		if !apiKeyIsSet {
			t.Fatal("apiKeyIsSet should be true after SetAPIKey")
		}
		if globalGeminiInstance == nil {
			t.Fatal("globalGeminiInstance should be initialized after SetAPIKey")
		}
		if globalGeminiInstance.apiKey != apiKeyForTest { // Assuming apiKey is stored in IsEvenAiGemini
			t.Fatalf("globalGeminiInstance.apiKey = %s; want %s", globalGeminiInstance.apiKey, apiKeyForTest)
		}

		// Test convenience functions
		resBool, errBool := IsEven(2)
		checkConvenienceResult(t, resBool, errBool, true, "IsEven", 2)
		// Add more checks as in original test
		resetGlobalStateAndClose() // Clean up for next test
	})
}

func TestConvenience_ApiKeyFromEnv_Gemini(t *testing.T) {
	resetGlobalStateAndClose()
	originalApiKey := os.Getenv("GEMINI_API_KEY") // Changed to GEMINI_API_KEY

	if originalApiKey == "" {
		t.Skip("Skipping TestConvenience_ApiKeyFromEnv_Gemini: GEMINI_API_KEY not set in environment.")
		return
	}

	err := SetAPIKey(originalApiKey)
	if err != nil {
		t.Fatalf("SetAPIKey with env key failed: %v", err)
	}

	resBool, errBool := IsEven(20)
	checkConvenienceResult(t, resBool, errBool, true, "IsEven", 20)
	resetGlobalStateAndClose()
}

func TestConvenience_NoAPIKeySet_Gemini(t *testing.T) {
	resetGlobalStateAndClose()

	_, err := IsEven(2)
	if err == nil {
		t.Fatal("Expected error when calling IsEven without API key, got nil")
	}
	expectedErrorMsg := "Gemini API key not set or instance not initialized. Call SetAPIKey() first." // Updated error message
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	err = SetAPIKey("")
	if err == nil {
		t.Fatal("Expected error when calling SetAPIKey with empty string, got nil")
	}
	if err.Error() != "API key cannot be empty" { // This error comes from SetAPIKey directly
		t.Errorf("Expected error 'API key cannot be empty', got '%s'", err.Error())
	}
	if apiKeyIsSet {
		t.Error("apiKeyIsSet should be false after SetAPIKey with empty string")
	}
	if globalGeminiInstance != nil {
		t.Error("globalGeminiInstance should be nil after SetAPIKey with empty string")
	}
}

func TestConvenience_SetAPIKeyWithModelOptions_Gemini(t *testing.T) {
	resetGlobalStateAndClose()
	apiKey := "test-key-for-options-gemini"
	if os.Getenv("GEMINI_API_KEY_FOR_TESTS") != "" {
		apiKey = os.Getenv("GEMINI_API_KEY_FOR_TESTS")
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	} else {
		t.Skip("Skipping TestConvenience_SetAPIKeyWithModelOptions_Gemini: No API key available.")
		return
	}

	customModel := "gemini-pro" // Example for Gemini
	var customTemp float32 = 0.7
	customOpts := GeminiModelOptions{Model: customModel, Temperature: &customTemp}
	err := SetAPIKey(apiKey, customOpts)
	if err != nil {
		t.Fatalf("SetAPIKey with custom model options failed: %v", err)
	}

	globalMu.Lock()
	defer globalMu.Unlock()

	if globalGeminiInstance == nil {
		t.Fatal("globalGeminiInstance is nil after SetAPIKey with custom options")
	}
	if globalGeminiInstance.genaiModel.ModelName != customOpts.Model {
		t.Errorf("Expected model %s, got %s", customOpts.Model, globalGeminiInstance.genaiModel.ModelName)
	}
	if globalGeminiInstance.genaiModel.GenerationConfig.Temperature == nil || *globalGeminiInstance.genaiModel.GenerationConfig.Temperature != *customOpts.Temperature {
		t.Errorf("Expected temperature %f, got %v", *customOpts.Temperature, globalGeminiInstance.genaiModel.GenerationConfig.Temperature)
	}
	resetGlobalStateAndClose()
}
