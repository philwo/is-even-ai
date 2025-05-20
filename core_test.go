package is_even_ai

import (
	"fmt"
	"strings"
	"testing"
)

// testPromptTemplates provides a set of mock prompt templates for testing.
var testPromptTemplates = IsEvenAiCorePromptTemplates{
	IsEven:        func(n int) string { return fmt.Sprintf("isEven %d", n) },
	IsOdd:         func(n int) string { return fmt.Sprintf("isOdd %d", n) },
	AreEqual:      func(a, b int) string { return fmt.Sprintf("areEqual %d %d", a, b) },
	AreNotEqual:   func(a, b int) string { return fmt.Sprintf("areNotEqual %d %d", a, b) },
	IsGreaterThan: func(a, b int) string { return fmt.Sprintf("isGreaterThan %d %d", a, b) },
	IsLessThan:    func(a, b int) string { return fmt.Sprintf("isLessThan %d %d", a, b) },
}

// mockQueryFunc is a mock implementation of QueryFunc for testing.
// It allows setting the expected return value and tracks calls.
type mockQueryFunc struct {
	called      bool
	lastPrompt  string
	returnValue *bool
	returnError error
}

func (m *mockQueryFunc) query(prompt string) (*bool, error) {
	m.called = true
	m.lastPrompt = prompt
	// Removed default true logic:
	// if m.returnValue == nil && m.returnError == nil { // Default to true if not set
	// 	defaultTrue := true
	// 	return &defaultTrue, nil
	// }
	return m.returnValue, m.returnError // Directly return what's set
}

func (m *mockQueryFunc) reset() {
	m.called = false
	m.lastPrompt = ""
	m.returnValue = nil
	m.returnError = nil
}

func TestIsEvenAiCore_DirectCalls(t *testing.T) {
	mockQuery := &mockQueryFunc{}

	core := NewIsEvenAiCore(testPromptTemplates, mockQuery.query)
	if core == nil {
		t.Fatal("NewIsEvenAiCore returned nil")
	}

	// Arguments for functions that take one int (isEven, isOdd)
	arg1 := 1
	// Arguments for functions that take two ints
	argA, argB := 1, 2

	testCases := []struct {
		name           string
		methodCall     func() (*bool, error)
		expectedPrompt string
		expectedResult bool
	}{
		{"IsEven", func() (*bool, error) { return core.IsEven(arg1) }, testPromptTemplates.IsEven(arg1), true},
		{"IsOdd", func() (*bool, error) { return core.IsOdd(arg1) }, testPromptTemplates.IsOdd(arg1), true},
		{"AreEqual", func() (*bool, error) { return core.AreEqual(argA, argB) }, testPromptTemplates.AreEqual(argA, argB), true},
		{"AreNotEqual", func() (*bool, error) { return core.AreNotEqual(argA, argB) }, testPromptTemplates.AreNotEqual(argA, argB), true},
		{"IsGreaterThan", func() (*bool, error) { return core.IsGreaterThan(argA, argB) }, testPromptTemplates.IsGreaterThan(argA, argB), true},
		{"IsLessThan", func() (*bool, error) { return core.IsLessThan(argA, argB) }, testPromptTemplates.IsLessThan(argA, argB), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery.reset()
			mockQuery.returnValue = &tc.expectedResult // Assume AI returns the expected for simplicity

			result, err := tc.methodCall()
			if err != nil {
				t.Fatalf("methodCall() for %s returned error: %v", tc.name, err)
			}
			if !mockQuery.called {
				t.Fatalf("QueryFunc was not called for %s", tc.name)
			}
			if mockQuery.lastPrompt != tc.expectedPrompt {
				t.Errorf("QueryFunc for %s called with wrong prompt. Got: '%s', Want: '%s'", tc.name, mockQuery.lastPrompt, tc.expectedPrompt)
			}
			if result == nil {
				t.Fatalf("Result for %s was nil, expected %t", tc.name, tc.expectedResult)
			}
			if *result != tc.expectedResult {
				t.Errorf("Result for %s was %t, expected %t", tc.name, *result, tc.expectedResult)
			}
		})
	}
}

func TestIsEvenAiCore_FallbackLogic(t *testing.T) {
	mockQuery := &mockQueryFunc{}

	// Create templates with optional ones missing
	partialTemplates := IsEvenAiCorePromptTemplates{
		IsEven:        testPromptTemplates.IsEven,
		AreEqual:      testPromptTemplates.AreEqual,
		IsGreaterThan: testPromptTemplates.IsGreaterThan,
		// IsOdd, AreNotEqual, IsLessThan are nil
	}

	core := NewIsEvenAiCore(partialTemplates, mockQuery.query)
	if core == nil {
		t.Fatal("NewIsEvenAiCore returned nil with partial templates")
	}

	arg1 := 1
	argA, argB := 1, 2

	// For fallback, the result is the negation of the complement's result
	// e.g., IsOdd falls back to !IsEven. If IsEven returns true, IsOdd should be false.
	aiReturnsTrue := true
	expectedFallbackResult := !aiReturnsTrue // If mock query (for complement) returns true, fallback is false

	testCases := []struct {
		name                string
		methodCall          func() (*bool, error)
		complementPromptGen func() string // Generates prompt for the complement method
		expectedResult      bool          // This is the final expected result after negation
	}{
		{
			name: "IsOdd (fallback to IsEven)",
			methodCall: func() (*bool, error) {
				return core.IsOdd(arg1)
			},
			complementPromptGen: func() string { return partialTemplates.IsEven(arg1) },
			expectedResult:      expectedFallbackResult,
		},
		{
			name: "AreNotEqual (fallback to AreEqual)",
			methodCall: func() (*bool, error) {
				return core.AreNotEqual(argA, argB)
			},
			complementPromptGen: func() string { return partialTemplates.AreEqual(argA, argB) },
			expectedResult:      expectedFallbackResult,
		},
		{
			name: "IsLessThan (fallback to IsGreaterThan)",
			methodCall: func() (*bool, error) {
				// IsLessThan(a, b) falls back to !IsGreaterThan(b, a)
				// So, the prompt for IsGreaterThan should use (argB, argA)
				return core.IsLessThan(argA, argB)
			},
			complementPromptGen: func() string { return partialTemplates.IsGreaterThan(argB, argA) },
			expectedResult:      expectedFallbackResult,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockQuery.reset()
			mockQuery.returnValue = &aiReturnsTrue // The complement AI call returns true

			result, err := tc.methodCall()
			if err != nil {
				t.Fatalf("methodCall() for %s returned error: %v", tc.name, err)
			}
			if !mockQuery.called {
				t.Fatalf("QueryFunc was not called for %s (during fallback)", tc.name)
			}

			expectedComplementPrompt := tc.complementPromptGen()
			if mockQuery.lastPrompt != expectedComplementPrompt {
				t.Errorf("QueryFunc for %s called with wrong complement prompt. Got: '%s', Want: '%s'", tc.name, mockQuery.lastPrompt, expectedComplementPrompt)
			}

			if result == nil {
				t.Fatalf("Result for %s was nil, expected %t", tc.name, tc.expectedResult)
			}
			if *result != tc.expectedResult {
				t.Errorf("Result for %s was %t, expected %t", tc.name, *result, tc.expectedResult)
			}
		})
	}
}

func TestNewIsEvenAiCore_NilQueryFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewIsEvenAiCore did not panic with nil query function")
		}
	}()
	NewIsEvenAiCore(testPromptTemplates, nil)
}

func TestIsEvenAiCore_GetPromptErrors(t *testing.T) {
	core := NewIsEvenAiCore(IsEvenAiCorePromptTemplates{}, func(prompt string) (*bool, error) { return nil, nil }) // Empty templates

	// Removed the following problematic check as it's covered by the mandatoryTemplates loop
	// and its assertion was expecting "not enough arguments" when "mandatory and not defined" is correct here.
	/*
		_, err := core.getPrompt("isEven") // Not enough args
		if err == nil || !strings.Contains(err.Error(), "not enough arguments") {
			t.Errorf("Expected error for not enough arguments for isEven, got %v", err)
		}
	*/

	// Test for mandatory templates not defined
	mandatoryTemplates := []string{"isEven", "areEqual", "isGreaterThan"}
	for _, mt := range mandatoryTemplates {
		t.Run(fmt.Sprintf("MandatoryTemplate_%s_Missing", mt), func(t *testing.T) {
			args := []int{1} // These args are for the prompt function if it were defined
			if mt == "areEqual" || mt == "isGreaterThan" {
				args = []int{1, 2}
			}
			// With empty templates, this will correctly error on the template being mandatory and not defined.
			_, err := core.getPrompt(mt, args...)
			if err == nil || !strings.Contains(err.Error(), "mandatory and not defined") {
				t.Errorf("Expected error for mandatory template %s not defined, got %v", mt, err)
			}
		})
	}

	_, err := core.getPrompt("unknownPrompt", 1)
	if err == nil || !strings.Contains(err.Error(), "unknown prompt name") {
		t.Errorf("Expected error for unknown prompt name, got %v", err)
	}

	// Add new sub-tests for "not enough arguments" when templates are defined
	t.Run("NotEnoughArguments", func(t *testing.T) {
		definedTemplates := IsEvenAiCorePromptTemplates{
			IsEven:        func(n int) string { return "isEven" },
			IsOdd:         func(n int) string { return "isOdd" },
			AreEqual:      func(a, b int) string { return "areEqual" },
			AreNotEqual:   func(a, b int) string { return "areNotEqual" },
			IsGreaterThan: func(a, b int) string { return "isGreaterThan" },
			IsLessThan:    func(a, b int) string { return "isLessThan" },
		}
		coreWithDefs := NewIsEvenAiCore(definedTemplates, func(prompt string) (*bool, error) { return nil, nil })

		argTestCases := []struct {
			name        string
			promptName  string
			args        []int
			expectedMsg string
		}{
			{"isEven_NoArgs", "isEven", []int{}, "not enough arguments for isEven prompt"},
			{"isOdd_NoArgs", "isOdd", []int{}, "not enough arguments for isOdd prompt"},
			{"areEqual_NoArgs", "areEqual", []int{}, "not enough arguments for areEqual prompt"},
			{"areEqual_OneArg", "areEqual", []int{1}, "not enough arguments for areEqual prompt"},
			{"areNotEqual_NoArgs", "areNotEqual", []int{}, "not enough arguments for areNotEqual prompt"},
			{"areNotEqual_OneArg", "areNotEqual", []int{1}, "not enough arguments for areNotEqual prompt"},
			{"isGreaterThan_NoArgs", "isGreaterThan", []int{}, "not enough arguments for isGreaterThan prompt"},
			{"isGreaterThan_OneArg", "isGreaterThan", []int{1}, "not enough arguments for isGreaterThan prompt"},
			{"isLessThan_NoArgs", "isLessThan", []int{}, "not enough arguments for isLessThan prompt"},
			{"isLessThan_OneArg", "isLessThan", []int{1}, "not enough arguments for isLessThan prompt"},
		}

		for _, tc := range argTestCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := coreWithDefs.getPrompt(tc.promptName, tc.args...)
				if err == nil || !strings.Contains(err.Error(), tc.expectedMsg) {
					t.Errorf("Expected error containing '%s', got %v", tc.expectedMsg, err)
				}
			})
		}
	})
}

func TestIsEvenAiCore_ErrorInQuery(t *testing.T) {
	mockQuery := &mockQueryFunc{}
	core := NewIsEvenAiCore(testPromptTemplates, mockQuery.query)
	queryError := fmt.Errorf("AI query failed")

	methods := map[string]func() (*bool, error){
		"IsEven":        func() (*bool, error) { return core.IsEven(1) },
		"IsOdd":         func() (*bool, error) { return core.IsOdd(1) },
		"AreEqual":      func() (*bool, error) { return core.AreEqual(1, 2) },
		"AreNotEqual":   func() (*bool, error) { return core.AreNotEqual(1, 2) },
		"IsGreaterThan": func() (*bool, error) { return core.IsGreaterThan(1, 2) },
		"IsLessThan":    func() (*bool, error) { return core.IsLessThan(1, 2) },
	}

	for name, methodCall := range methods {
		t.Run(name+"_QueryError", func(t *testing.T) {
			mockQuery.reset()
			mockQuery.returnError = queryError

			_, err := methodCall()
			if err == nil {
				t.Errorf("Expected error from %s, got nil", name)
			} else if !strings.Contains(err.Error(), queryError.Error()) {
				t.Errorf("Expected error from %s to contain '%s', got '%s'", name, queryError.Error(), err.Error())
			}
		})
	}
}

func TestIsEvenAiCore_UndefinedResponse(t *testing.T) {
	mockQuery := &mockQueryFunc{} // returnValue will be nil by default in reset
	core := NewIsEvenAiCore(testPromptTemplates, mockQuery.query)

	methods := map[string]func() (*bool, error){
		"IsEven":        func() (*bool, error) { return core.IsEven(1) },
		"IsOdd":         func() (*bool, error) { return core.IsOdd(1) },
		"AreEqual":      func() (*bool, error) { return core.AreEqual(1, 2) },
		"AreNotEqual":   func() (*bool, error) { return core.AreNotEqual(1, 2) },
		"IsGreaterThan": func() (*bool, error) { return core.IsGreaterThan(1, 2) },
		"IsLessThan":    func() (*bool, error) { return core.IsLessThan(1, 2) },
	}

	for name, methodCall := range methods {
		t.Run(name+"_Undefined", func(t *testing.T) {
			mockQuery.reset() // Ensures returnValue is nil

			res, err := methodCall()
			if err != nil {
				t.Errorf("Expected no error for undefined response from %s, got %v", name, err)
			}
			if res != nil {
				t.Errorf("Expected nil result for undefined response from %s, got %v", name, *res)
			}
		})
	}
}

func TestIsEvenAiCore_FallbackUndefinedResponse(t *testing.T) {
	mockQuery := &mockQueryFunc{} // returnValue will be nil by default in reset
	partialTemplates := IsEvenAiCorePromptTemplates{
		IsEven:        testPromptTemplates.IsEven,
		AreEqual:      testPromptTemplates.AreEqual,
		IsGreaterThan: testPromptTemplates.IsGreaterThan,
	}
	core := NewIsEvenAiCore(partialTemplates, mockQuery.query)

	methods := map[string]func() (*bool, error){
		"IsOdd_Fallback_Undefined":       func() (*bool, error) { return core.IsOdd(1) },
		"AreNotEqual_Fallback_Undefined": func() (*bool, error) { return core.AreNotEqual(1, 2) },
		"IsLessThan_Fallback_Undefined":  func() (*bool, error) { return core.IsLessThan(1, 2) },
	}

	for name, methodCall := range methods {
		t.Run(name, func(t *testing.T) {
			mockQuery.reset() // Ensures returnValue is nil

			res, err := methodCall()
			if err != nil {
				t.Errorf("Expected no error for fallback undefined response from %s, got %v", name, err)
			}
			if res != nil {
				t.Errorf("Expected nil result for fallback undefined response from %s, got %v", name, *res)
			}
		})
	}
}
