package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"bufio"
	"strings"

	agk "github.com/agenticgokit/agenticgokit/v1beta"
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/openai"
)

func main() {
	ctx := context.Background()
	// Get service version from environment or use default
	serviceVersion := os.Getenv("SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "0.1.0"
	}

	// Create a simple agent with automatic observability configuration
	// Tracing is enabled if AGK_TRACE=true environment variable is set
	// Usage: AGK_TRACE=true go run main.go
	agent, err := agk.NewBuilder("agent-2").
		WithLLM("ollama", "llama3.2:1b").
		WithObservability("agent-2", serviceVersion).
		Build()
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Cleanup(ctx)

	// Take User input at runtime
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n💬 Chat started (type 'exit' to quit)\n")

	for{
		reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		fmt.Print("\n\nYou: ")
		userMessage, _ := reader.ReadString('\n')
		userMessage = strings.TrimSpace(userMessage)
		
		if userMessage == "exit"{
			fmt.Println("👋 Goodbye!")
			cancel()
			break
		}
		fmt.Println("\nAssistant: ")

		stream, err := agent.RunStream(reqCtx, userMessage)
		if err != nil {
			log.Printf("Failed to start streaming: %v", err)
			cancel()
			continue
		}

		printStreamingResponse(stream)
		fmt.Println()
		cancel()
	}
}

// printStreamingResponse prints the streaming response as tokens arrive
func printStreamingResponse(stream agk.Stream) {
	for chunk := range stream.Chunks() {
		if chunk.Error != nil {
			fmt.Printf("\n❌ Error: %v\n", chunk.Error)
			break
		}

		switch chunk.Type {
		case agk.ChunkTypeDelta:
			fmt.Print(chunk.Delta)
		case agk.ChunkTypeDone:
			fmt.Println("\n\n✅ Completed")
		}
	}
}
