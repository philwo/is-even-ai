package main

import (
	"fmt"
	"log"
	"os" // For API Key from environment variable

	isevenai "github.com/philwo/is-even-ai"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY") // Changed from OPENAI_API_KEY
	if apiKey == "" {
		log.Fatal("Error: GEMINI_API_KEY environment variable not set.")
	}

	// Set the API key for the global instance
	// Optionally, pass custom model options:
	// var temp float32 = 0.1
	// customModelOpts := isevenai.GeminiModelOptions{Model: "gemini-pro", Temperature: &temp}
	// err := isevenai.SetAPIKey(apiKey, customModelOpts)
	err := isevenai.SetAPIKey(apiKey)
	if err != nil {
		log.Fatalf("Error setting API key: %v", err)
	}

	// --- Using convenience functions ---
	fmt.Println("Using convenience functions:")

	// Define a slice of numbers to test for IsEven and IsOdd
	numbers := []int{4, 7, 0, 13, -2, 42} // Example numbers

	for _, num := range numbers {
		// Check if num is even
		isEvenResult, err := isevenai.IsEven(num)
		if err != nil {
			log.Printf("Error checking if %d is even: %v", num, err)
		} else {
			if isEvenResult == nil {
				fmt.Printf("Is %d even? Undefined\n", num)
			} else {
				fmt.Printf("Is %d even? %t\n", num, *isEvenResult)
			}
		}

		// Check if num is odd
		isOddResult, err := isevenai.IsOdd(num)
		if err != nil {
			log.Printf("Error checking if %d is odd: %v", num, err)
		} else {
			if isOddResult == nil {
				fmt.Printf("Is %d odd? Undefined\n", num)
			} else {
				fmt.Printf("Is %d odd? %t\n", num, *isOddResult)
			}
		}
		fmt.Println() // Add a blank line to separate results for different numbers
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
	clientOptions := isevenai.GeminiClientOptions{APIKey: apiKey}
	// Default model is gemini-2.0-flash-lite, temp 0.
	// To customize:
	// var temp float32 = 0.2
	// modelOpts := isevenai.GeminiModelOptions{Model: "gemini-pro", Temperature: &temp}
	// myAiInstance, err := isevenai.NewIsEvenAiGemini(clientOptions, modelOpts)
	myAiInstance, err := isevenai.NewIsEvenAiGemini(clientOptions)
	if err != nil {
		log.Fatalf("Error creating direct IsEvenAiGemini instance: %v", err)
	}
	defer myAiInstance.Close() // Close the client when done

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
