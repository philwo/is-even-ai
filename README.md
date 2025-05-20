# is-even-ai-gemini (Go Version)

[![Go Reference](https://pkg.go.dev/badge/github.com/philwo/is-even-ai.svg)](https://pkg.go.dev/github.com/philwo/is-even-ai)
[![LICENSE](https://img.shields.io/github/license/philwo/is-even-ai.svg?style=flat)](https://github.com/philwo/is-even-ai/blob/main/LICENSE)


Check if a number is even using the power of ✨AI✨ with Google Gemini.

Uses Google's Gemini AI models (defaulting to `gemini-2.0-flash-lite`) under the hood to determine if a number is even, odd, equal, etc.

For all those who want to use AI in their Go product but don't know how.

Inspired by the famous [`is-even`](https://www.npmjs.com/package/is-even) npm package and related AI adaptations. This is a Go adaptation.

## Installation

```sh
go get [github.com/philwo/is-even-ai](https://github.com/philwo/is-even-ai)

## Usage

First, ensure you have a Gemini API key. You can set it as an environment variable `GEMINI_API_KEY`.

### Convenience Functions

```go
package main

import (
	"fmt"
	"log"
	"os"

	isevenai "[github.com/philwo/is-even-ai](https://github.com/philwo/is-even-ai)"
)

func main() {
	// Set API Key (reads from GEMINI_API_KEY environment variable in this example)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY environment variable not set.")
	}
	err := isevenai.SetAPIKey(apiKey)
	if err != nil {
		log.Fatalf("Error setting API key: %v", err)
	}

	fmt.Println(isevenai.IsEven(2))    // &true, <nil>
	fmt.Println(isevenai.IsEven(3))    // &false, <nil>
	fmt.Println(isevenai.IsOdd(4))     // &false, <nil>
	fmt.Println(isevenai.IsOdd(5))     // &true, <nil>
	fmt.Println(isevenai.AreEqual(6, 6)) // &true, <nil>
	// ... and so on for AreNotEqual, IsGreaterThan, IsLessThan
}
```

### Direct Instance Usage

For more advanced usage, like changing which model to use or setting the temperature, use `IsEvenAiGemini` directly.

```go
package main

import (
	"fmt"
	"log"
	"os"

	isevenai "[github.com/philwo/is-even-ai](https://github.com/philwo/is-even-ai)"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY environment variable not set.")
	}

	clientOpts := isevenai.GeminiClientOptions{
		APIKey: apiKey,
	}

	// Optional: Customize model and temperature
	// var temp float32 = 0.5
	// modelOpts := isevenai.GeminiModelOptions{
	// 	Model:       "gemini-pro", // Example: use gemini-pro
	// 	Temperature: &temp,
	// }
	// geminiAI, err := isevenai.NewIsEvenAiGemini(clientOpts, modelOpts)

	geminiAI, err := isevenai.NewIsEvenAiGemini(clientOpts) // Uses gemini-2.0-flash-lite by default
	if err != nil {
		log.Fatalf("Failed to create IsEvenAiGemini instance: %v", err)
	}
	defer geminiAI.Close() // Important to close the client

	result, err := geminiAI.IsEven(2)
	if err != nil {
		log.Printf("Error: %v", err)
	} else if result != nil {
		fmt.Printf("Is 2 even? %t\n", *result) // Is 2 even? true
	} else {
		fmt.Println("Is 2 even? Undefined")
	}
	// ... other method calls
}
```

## Supported AI platforms

- [x] Google Gemini via `IsEvenAiGemini` (using `gemini-2.0-flash-lite` by default)

## Supported methods

The following methods return `(*bool, error)`. The `*bool` can be true, false, or nil (if the AI's response is undefined).

- `IsEven(n int)`
- `IsOdd(n int)`
- `AreEqual(a int, b int)`
- `AreNotEqual(a int, b int)`
- `IsGreaterThan(a int, b int)`
- `IsLessThan(a int, b int)`
