// Copyright 2025 Google LLC

// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file or at https://opensource.org/licenses/MIT.

package is_even_ai

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// Helper function to check boolean pointer results for Gemini tests
func checkGeminiResult(t *testing.T, val *bool, err error, expected bool, funcName string, inputs ...int) {
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

func TestIsEvenAiGemini_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping Gemini integration tests: GEMINI_API_KEY not set")
	}

	clientOpts := GeminiClientOptions{APIKey: apiKey}
	// Default model is gemini-2.0-flash-lite, temperature 0
	ai, err := NewIsEvenAiGemini(clientOpts)
	if err != nil {
		t.Fatalf("Failed to create NewIsEvenAiGemini: %v", err)
	}
	defer func() { _ = ai.Close() }() // Checked error

	t.Run("IsEven", func(t *testing.T) {
		res, err := ai.IsEven(2)
		checkGeminiResult(t, res, err, true, "IsEven", 2)
		res, err = ai.IsEven(3)
		checkGeminiResult(t, res, err, false, "IsEven", 3)
	})

	t.Run("IsOdd", func(t *testing.T) {
		res, err := ai.IsOdd(4)
		checkGeminiResult(t, res, err, false, "IsOdd", 4)
		res, err = ai.IsOdd(5)
		checkGeminiResult(t, res, err, true, "IsOdd", 5)
	})

	t.Run("AreEqual", func(t *testing.T) {
		res, err := ai.AreEqual(6, 6)
		checkGeminiResult(t, res, err, true, "AreEqual", 6, 6)
		res, err = ai.AreEqual(6, 7)
		checkGeminiResult(t, res, err, false, "AreEqual", 6, 7)
	})

	t.Run("AreNotEqual", func(t *testing.T) {
		res, err := ai.AreNotEqual(6, 7)
		checkGeminiResult(t, res, err, true, "AreNotEqual", 6, 7)
		res, err = ai.AreNotEqual(7, 7)
		checkGeminiResult(t, res, err, false, "AreNotEqual", 7, 7)
	})

	t.Run("IsGreaterThan", func(t *testing.T) {
		res, err := ai.IsGreaterThan(8, 7)
		checkGeminiResult(t, res, err, true, "IsGreaterThan", 8, 7)
		res, err = ai.IsGreaterThan(7, 8)
		checkGeminiResult(t, res, err, false, "IsGreaterThan", 7, 8)
	})

	t.Run("IsLessThan", func(t *testing.T) {
		res, err := ai.IsLessThan(8, 9)
		checkGeminiResult(t, res, err, true, "IsLessThan", 8, 9)
		res, err = ai.IsLessThan(9, 8)
		checkGeminiResult(t, res, err, false, "IsLessThan", 9, 8)
	})
}

func TestNewIsEvenAiGemini_Options(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping Gemini options test: GEMINI_API_KEY not set")
	}

	t.Run("DefaultOptions", func(t *testing.T) {
		clientOpts := GeminiClientOptions{APIKey: apiKey}
		ai, err := NewIsEvenAiGemini(clientOpts)
		if err != nil {
			t.Fatalf("NewIsEvenAiGemini failed: %v", err)
		}
		defer func() { _ = ai.Close() }() // Checked error

		if ai.modelName != "gemini-2.0-flash-lite" {
			t.Errorf("Expected default model gemini-2.0-flash-lite, got %s", ai.modelName)
		}
		// QF1008: Use direct access to Temperature due to embedding
		if ai.genaiModel.Temperature == nil || *ai.genaiModel.Temperature != 0.0 {
			temp := "nil"
			if ai.genaiModel.Temperature != nil {
				temp = fmt.Sprintf("%f", *ai.genaiModel.Temperature)
			}
			t.Errorf("Expected default temperature 0.0, got %s", temp)
		}
	})

	t.Run("CustomModelOptions", func(t *testing.T) {
		clientOpts := GeminiClientOptions{APIKey: apiKey}
		customModel := "gemini-pro" // Example custom model
		var customTemp float32 = 0.7
		modelOpts := GeminiModelOptions{Model: customModel, Temperature: &customTemp}

		ai, err := NewIsEvenAiGemini(clientOpts, modelOpts)
		if err != nil {
			t.Fatalf("NewIsEvenAiGemini failed: %v", err)
		}
		defer func() { _ = ai.Close() }() // Checked error

		if ai.modelName != customModel {
			t.Errorf("Expected custom model %s, got %s", customModel, ai.modelName)
		}
		// QF1008: Use direct access to Temperature due to embedding
		if ai.genaiModel.Temperature == nil || *ai.genaiModel.Temperature != customTemp {
			temp := "nil"
			if ai.genaiModel.Temperature != nil {
				temp = fmt.Sprintf("%f", *ai.genaiModel.Temperature)
			}
			t.Errorf("Expected custom temperature %f, got %s", customTemp, temp)
		}
	})

	t.Run("EmptyAPIKey", func(t *testing.T) {
		_, err := NewIsEvenAiGemini(GeminiClientOptions{APIKey: ""})
		if err == nil {
			t.Error("Expected error for empty API key, got nil")
		} else if err.Error() != "gemini API key is required" { // ST1005: uncapitalized
			t.Errorf("Expected error 'gemini API key is required', got '%s'", err.Error())
		}
	})
}

func TestIsEvenAiGemini_APIFailure(t *testing.T) {
	// Using an invalid API key should cause an error during the API call.
	clientOpts := GeminiClientOptions{APIKey: "invalid-gemini-api-key-for-test"}
	ai, err := NewIsEvenAiGemini(clientOpts)
	if err != nil {
		t.Fatalf("NewIsEvenAiGemini with invalid key unexpectedly failed on creation: %v (expected failure on call)", err)
	}
	defer func() { _ = ai.Close() }() // Checked error

	_, err = ai.IsEven(2)
	if err == nil {
		t.Error("Expected an error from IsEven call with failing API key, got nil")
	} else {
		t.Logf("Got expected error (this is good for this test case): %v", err)
		if !strings.Contains(strings.ToLower(err.Error()), "api key not valid") &&
			!strings.Contains(strings.ToLower(err.Error()), "permission denied") &&
			!strings.Contains(strings.ToLower(err.Error()), "authentication") {
			t.Logf("Warning: Error message '%s' does not explicitly state API key invalidity or permission issue, but an error was correctly raised.", err.Error())
		}
	}
}
