package is_even_ai

import (
	"errors"
	"fmt"
)

// PromptTemplate1 defines a function that takes one integer argument and returns a string prompt.
type PromptTemplate1 func(n int) string

// PromptTemplate2 defines a function that takes two integer arguments and returns a string prompt.
type PromptTemplate2 func(a, b int) string

// IsEvenAiCorePromptTemplates holds the templates for generating prompts.
//   - IsEven, AreEqual, IsGreaterThan are mandatory.
//   - IsOdd, AreNotEqual, IsLessThan are optional. If a template for an optional
//     operation is nil, the corresponding method will use a fallback strategy
//     (e.g., IsOdd will be derived from !IsEven).
//
// All prompt template functions are synchronous and return a string.
type IsEvenAiCorePromptTemplates struct {
	IsEven        PromptTemplate1
	IsOdd         PromptTemplate1 // Optional: if nil, IsOdd will be derived from !IsEven
	AreEqual      PromptTemplate2
	AreNotEqual   PromptTemplate2 // Optional: if nil, AreNotEqual will be derived from !AreEqual
	IsGreaterThan PromptTemplate2
	IsLessThan    PromptTemplate2 // Optional: if nil, IsLessThan will be derived from !IsGreaterThan(b,a)
}

// QueryFunc defines a function that takes a prompt string, queries an AI model,
// and returns a boolean result or an error. The *bool type allows for true, false,
// or nil (representing an undefined or indeterminate answer from the AI).
type QueryFunc func(prompt string) (result *bool, err error)

// IsEvenAiCore provides the core functionality for querying number properties using AI.
type IsEvenAiCore struct {
	promptTemplates IsEvenAiCorePromptTemplates
	query           QueryFunc
}

// NewIsEvenAiCore creates a new instance of IsEvenAiCore.
// It requires a set of prompt templates and a query function to interact with an AI.
func NewIsEvenAiCore(templates IsEvenAiCorePromptTemplates, query QueryFunc) *IsEvenAiCore {
	if query == nil {
		panic("query function cannot be nil") // Or return an error
	}
	return &IsEvenAiCore{
		promptTemplates: templates,
		query:           query,
	}
}

// getPrompt retrieves and formats a prompt string based on the prompt name and arguments.
// For optional templates that are not provided, it returns an empty string and no error.
func (c *IsEvenAiCore) getPrompt(promptName string, args ...int) (string, error) {
	switch promptName {
	case "isEven":
		if c.promptTemplates.IsEven == nil {
			return "", errors.New("isEven prompt template is mandatory and not defined")
		}
		if len(args) < 1 {
			return "", errors.New("not enough arguments for isEven prompt")
		}
		return c.promptTemplates.IsEven(args[0]), nil
	case "isOdd":
		if c.promptTemplates.IsOdd == nil {
			return "", nil // Optional, return empty string if not defined
		}
		if len(args) < 1 {
			return "", errors.New("not enough arguments for isOdd prompt")
		}
		return c.promptTemplates.IsOdd(args[0]), nil
	case "areEqual":
		if c.promptTemplates.AreEqual == nil {
			return "", errors.New("areEqual prompt template is mandatory and not defined")
		}
		if len(args) < 2 {
			return "", errors.New("not enough arguments for areEqual prompt")
		}
		return c.promptTemplates.AreEqual(args[0], args[1]), nil
	case "areNotEqual":
		if c.promptTemplates.AreNotEqual == nil {
			return "", nil // Optional
		}
		if len(args) < 2 {
			return "", errors.New("not enough arguments for areNotEqual prompt")
		}
		return c.promptTemplates.AreNotEqual(args[0], args[1]), nil
	case "isGreaterThan":
		if c.promptTemplates.IsGreaterThan == nil {
			return "", errors.New("isGreaterThan prompt template is mandatory and not defined")
		}
		if len(args) < 2 {
			return "", errors.New("not enough arguments for isGreaterThan prompt")
		}
		return c.promptTemplates.IsGreaterThan(args[0], args[1]), nil
	case "isLessThan":
		if c.promptTemplates.IsLessThan == nil {
			return "", nil // Optional
		}
		if len(args) < 2 {
			return "", errors.New("not enough arguments for isLessThan prompt")
		}
		return c.promptTemplates.IsLessThan(args[0], args[1]), nil
	default:
		return "", fmt.Errorf("unknown prompt name: %s", promptName)
	}
}

// IsEven checks if a number 'n' is even.
// Returns a pointer to boolean (*bool) and an error.
// *bool can be true, false, or nil (if the AI's response is undefined).
func (c *IsEvenAiCore) IsEven(n int) (*bool, error) {
	prompt, err := c.getPrompt("isEven", n)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt for IsEven: %w", err)
	}
	return c.query(prompt)
}

// IsOdd checks if a number 'n' is odd.
// If an 'isOdd' prompt template is not provided, it derives the result by negating IsEven(n).
func (c *IsEvenAiCore) IsOdd(n int) (*bool, error) {
	prompt, err := c.getPrompt("isOdd", n)
	if err != nil {
		// This error means getPrompt failed (e.g., not enough args for a defined template,
		// or a misconfiguration). It should not proceed to fallback.
		return nil, fmt.Errorf("failed to get prompt for IsOdd: %w", err)
	}

	if prompt != "" { // Template was provided and prompt generated successfully
		return c.query(prompt)
	}

	// Fallback: template was optional and not provided (i.e., prompt == "" and err == nil from getPrompt)
	isEvenResult, err := c.IsEven(n)
	if err != nil {
		return nil, fmt.Errorf("failed to determine IsOdd by inverting IsEven: %w", err)
	}
	if isEvenResult == nil { // IsEven returned undefined
		return nil, nil
	}
	res := !(*isEvenResult)
	return &res, nil
}

// AreEqual checks if numbers 'a' and 'b' are equal.
func (c *IsEvenAiCore) AreEqual(a, b int) (*bool, error) {
	prompt, err := c.getPrompt("areEqual", a, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt for AreEqual: %w", err)
	}
	return c.query(prompt)
}

// AreNotEqual checks if numbers 'a' and 'b' are not equal.
// If an 'areNotEqual' prompt template is not provided, it derives the result by negating AreEqual(a,b).
func (c *IsEvenAiCore) AreNotEqual(a, b int) (*bool, error) {
	prompt, err := c.getPrompt("areNotEqual", a, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt for AreNotEqual: %w", err)
	}

	if prompt != "" { // Template was provided and prompt generated successfully
		return c.query(prompt)
	}

	// Fallback: template was optional and not provided
	areEqualResult, err := c.AreEqual(a, b)
	if err != nil {
		return nil, fmt.Errorf("failed to determine AreNotEqual by inverting AreEqual: %w", err)
	}
	if areEqualResult == nil { // AreEqual returned undefined
		return nil, nil
	}
	res := !(*areEqualResult)
	return &res, nil
}

// IsGreaterThan checks if number 'a' is greater than number 'b'.
func (c *IsEvenAiCore) IsGreaterThan(a, b int) (*bool, error) {
	prompt, err := c.getPrompt("isGreaterThan", a, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt for IsGreaterThan: %w", err)
	}
	return c.query(prompt)
}

// IsLessThan checks if number 'a' is less than number 'b'.
// If an 'isLessThan' prompt template is not provided, it derives the result by checking !IsGreaterThan(b,a).
func (c *IsEvenAiCore) IsLessThan(a, b int) (*bool, error) {
	prompt, err := c.getPrompt("isLessThan", a, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt for IsLessThan: %w", err)
	}

	if prompt != "" { // Template was provided and prompt generated successfully
		return c.query(prompt)
	}

	// Fallback: template was optional and not provided
	isGreaterThanResult, err := c.IsGreaterThan(b, a) // Note: arguments are swapped
	if err != nil {
		return nil, fmt.Errorf("failed to determine IsLessThan by inverting IsGreaterThan(b,a): %w", err)
	}
	if isGreaterThanResult == nil { // IsGreaterThan(b,a) returned undefined
		return nil, nil
	}
	res := !(*isGreaterThanResult)
	return &res, nil
}
