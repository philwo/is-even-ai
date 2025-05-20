// Copyright 2025 Google LLC

// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file or at https://opensource.org/licenses/MIT.

package is_even_ai

import (
	"os"
	"testing"
)

// Helper to reset global state for convenience tests
func resetGlobalStateAndClose() {
	globalMu.Lock()
	if globalGeminiInstance != nil {
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
	t.Cleanup(resetGlobalStateAndClose)

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
		if originalApiKey != "" {
			currentEnvKey := os.Getenv("GEMINI_API_KEY")
			_ = os.Unsetenv("GEMINI_API_KEY")
			defer func() { _ = os.Setenv("GEMINI_API_KEY", currentEnvKey) }()
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
	})
}

func TestConvenience_ApiKeyFromEnv_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose)

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
}

func TestConvenience_NoAPIKeySet_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose)

	_, err := IsEven(2)
	if err == nil {
		t.Fatal("Expected error when calling IsEven without API key, got nil")
	}
	expectedErrorMsg := "gemini API key not set or instance not initialized. Call SetAPIKey() first"
	if err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
	}

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
	if globalGeminiInstance != nil {
		t.Error("globalGeminiInstance should be nil after SetAPIKey with empty string")
	}
}

func TestConvenience_SetAPIKeyWithModelOptions_Gemini(t *testing.T) {
	t.Cleanup(resetGlobalStateAndClose)

	var apiKey string
	envKeyForTests := os.Getenv("GEMINI_API_KEY_FOR_TESTS")
	envKeyGlobal := os.Getenv("GEMINI_API_KEY")

	if envKeyForTests != "" {
		apiKey = envKeyForTests
	} else if envKeyGlobal != "" {
		apiKey = envKeyGlobal
	} else {
		t.Skip("Skipping TestConvenience_SetAPIKeyWithModelOptions_Gemini: No API key available from GEMINI_API_KEY_FOR_TESTS or GEMINI_API_KEY.")
		return
	}

	customModel := "gemini-pro"
	var customTemp float32 = 0.7
	customOpts := GeminiModelOptions{Model: customModel, Temperature: &customTemp}
	err := SetAPIKey(apiKey, customOpts)
	if err != nil {
		t.Fatalf("SetAPIKey with custom model options failed: %v", err)
	}

	globalMu.Lock()
	instanceToCheck := globalGeminiInstance
	globalMu.Unlock()

	if instanceToCheck == nil {
		t.Fatal("globalGeminiInstance is nil after SetAPIKey with custom options")
	}
	if instanceToCheck.modelName != customOpts.Model {
		t.Errorf("Expected model %s, got %s", customOpts.Model, instanceToCheck.modelName)
	}
	if instanceToCheck.genaiModel.Temperature == nil || *instanceToCheck.genaiModel.Temperature != *customOpts.Temperature {
		t.Errorf("Expected temperature %f, got %v", *customOpts.Temperature, instanceToCheck.genaiModel.Temperature)
	}
}
