package main

import (
	"fmt"
	"log"
	"os" // For API Key from environment variable

	isevenai "github.com/philwo/is-even-ai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable not set.")
	}

	// Set the API key for the global instance
	// Optionally, pass custom chat options:
	// customChatOpts := isevenai.OpenAIChatOptions{Model: "gpt-4", Temperature: 0.1}
	// err := isevenai.SetAPIKey(apiKey, customChatOpts)
	err := isevenai.SetAPIKey(apiKey)
	if err != nil {
		log.Fatalf("Error setting API key: %v", err)
	}

	// --- Using convenience functions ---
	fmt.Println("Using convenience functions:")

	num1 := 4
	isEvenResult, err := isevenai.IsEven(num1)
	if err != nil {
		log.Printf("Error checking if %d is even: %v", num1, err)
	} else {
		if isEvenResult == nil {
			fmt.Printf("Is %d even? Undefined\n", num1)
		} else {
			fmt.Printf("Is %d even? %t\n", num1, *isEvenResult)
		}
	}

	num2 := 7
	isOddResult, err := isevenai.IsOdd(num2)
	if err != nil {
		log.Printf("Error checking if %d is odd: %v", num2, err)
	} else {
		if isOddResult == nil {
			fmt.Printf("Is %d odd? Undefined\n", num2)
		} else {
			fmt.Printf("Is %d odd? %t\n", num2, *isOddResult)
		}
	}

	valA, valB := 10, 10
	areEqualResult, err := isevenai.AreEqual(valA, valB)
	if err != nil {
		log.Printf("Error checking if %d and %d are equal: %v", valA, valB, err)
	} else {
		if areEqualResult == nil {
			fmt.Printf("Are %d and %d equal? Undefined\n", valA, valB)
		} else {
			fmt.Printf("Are %d and %d equal? %t\n", valA, valB, *areEqualResult)
		}
	}

	// --- Alternatively, creating an instance directly ---
	// This is useful if you don't want to use the global instance or need multiple instances.
	fmt.Println("\nUsing a direct instance:")
	clientOptions := isevenai.OpenAIClientOptions{APIKey: apiKey}
	myAiInstance, err := isevenai.NewIsEvenAiOpenAi(clientOptions)
	if err != nil {
		log.Fatalf("Error creating direct IsEvenAiOpenAi instance: %v", err)
	}

	isNumGreaterThan, err := myAiInstance.IsGreaterThan(100, 50)
	if err != nil {
		log.Printf("Error checking IsGreaterThan with direct instance: %v", err)
	} else {
		if isNumGreaterThan == nil {
			fmt.Println("Is 100 > 50? Undefined")
		} else {
			fmt.Printf("Is 100 > 50? %t\n", *isNumGreaterThan)
		}
	}
}
