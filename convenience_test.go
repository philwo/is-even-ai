package is_even_ai

import (
	"os"
	"testing"
)

// Helper to reset global state for convenience tests
func resetGlobalStateAndClose() {
	globalMu.Lock()
	if globalGeminiInstance != nil {
		// Consider logging an error from Close() if it occurs during cleanup
		_ = globalGeminiInstance.Close() // Best effort close
		globalGeminiInstance = nil
	}
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

func TestConvenience_SetAPIKeyAndUse_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose) // Ensure cleanup after this test

	originalApiKey := os.Getenv("GEMINI_API_KEY")
	apiKeyForTest := "test-api-key-from-setapikey-gemini"

	if originalApiKey == "" {
		t.Log("GEMINI_API_KEY not in env, SetAPIKey test will use a dummy key for instantiation logic only.")
		if os.Getenv("GEMINI_API_KEY_FOR_TESTS") != "" {
			apiKeyForTest = os.Getenv("GEMINI_API_KEY_FOR_TESTS")
		} else {
			t.Skip("Skipping TestConvenience_SetAPIKeyAndUse_Gemini: GEMINI_API_KEY_FOR_TESTS not set, and no fallback for full convenience function test without a real key.")
			return
		}
	} else {
		apiKeyForTest = originalApiKey
	}

	t.Run("WithKeyPassedToSetAPIKey_Gemini", func(t *testing.T) {
		// resetGlobalStateAndClose() // Already handled by t.Cleanup at parent level, or can be specific if sub-test manipulates state uniquely
		// For subtests, if they independently manipulate the global state and need reset before *other subtests*,
		// then having reset here is also fine. Or ensure parent cleanup is sufficient.
		// Let's keep it simple: the parent test's Cleanup should cover this.
		// If a sub-test fails and SetAPIKey was called, parent cleanup handles it.

		if originalApiKey != "" {
			currentEnvKey := os.Getenv("GEMINI_API_KEY")
			os.Unsetenv("GEMINI_API_KEY")
			defer os.Setenv("GEMINI_API_KEY", currentEnvKey)
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
		if globalGeminiInstance.apiKey != apiKeyForTest {
			t.Fatalf("globalGeminiInstance.apiKey = %s; want %s", globalGeminiInstance.apiKey, apiKeyForTest)
		}

		resBool, errBool := IsEven(2)
		checkConvenienceResult(t, resBool, errBool, true, "IsEven", 2)
		// resetGlobalStateAndClose() // Let t.Cleanup handle final state
	})
}

func TestConvenience_ApiKeyFromEnv_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose) // Ensure cleanup after this test
	// resetGlobalStateAndClose() // No longer needed at start if t.Cleanup is used

	originalApiKey := os.Getenv("GEMINI_API_KEY")

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
	// resetGlobalStateAndClose() // Let t.Cleanup handle final state
}

func TestConvenience_NoAPIKeySet_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose) // Ensure cleanup after this test
	// resetGlobalStateAndClose() // No longer needed at start

	_, err := IsEven(2)
	if err == nil {
		t.Fatal("Expected error when calling IsEven without API key, got nil")
	}
	expectedErrorMsg := "Gemini API key not set or instance not initialized. Call SetAPIKey() first."
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

	err = SetAPIKey("") // This will attempt to clear/reset the global instance
	if err == nil {
		t.Fatal("Expected error when calling SetAPIKey with empty string, got nil")
	}
	if err.Error() != "API key cannot be empty" {
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
	t.Cleanup(resetGlobalStateAndClose) // Ensure cleanup after this test
	// resetGlobalStateAndClose() // No longer needed at start

	apiKey := "test-key-for-options-gemini"
	if os.Getenv("GEMINI_API_KEY_FOR_TESTS") != "" {
		apiKey = os.Getenv("GEMINI_API_KEY_FOR_TESTS")
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	} else {
		t.Skip("Skipping TestConvenience_SetAPIKeyWithModelOptions_Gemini: No API key available.")
		return
	}

	customModel := "gemini-pro"
	var customTemp float32 = 0.7
	customOpts := GeminiModelOptions{Model: customModel, Temperature: &customTemp}
	err := SetAPIKey(apiKey, customOpts)
	if err != nil {
		// If the hang was due to NewClient, it should now timeout and fail here.
		t.Fatalf("SetAPIKey with custom model options failed: %v", err)
	}

	// Lock here to check global state is dangerous if SetAPIKey failed to unlock,
	// but SetAPIKey uses defer for its lock.
	globalMu.Lock()
	instanceToCheck := globalGeminiInstance // Copy the value under lock
	globalMu.Unlock()                       // Unlock immediately after reading

	if instanceToCheck == nil {
		t.Fatal("globalGeminiInstance is nil after SetAPIKey with custom options")
	}
	if instanceToCheck.modelName != customOpts.Model {
		t.Errorf("Expected model %s, got %s", customOpts.Model, instanceToCheck.modelName)
	}
	if instanceToCheck.genaiModel.GenerationConfig.Temperature == nil || *instanceToCheck.genaiModel.GenerationConfig.Temperature != *customOpts.Temperature {
		t.Errorf("Expected temperature %f, got %v", *customOpts.Temperature, instanceToCheck.genaiModel.GenerationConfig.Temperature)
	}
	// resetGlobalStateAndClose() // Let t.Cleanup handle final state
}
