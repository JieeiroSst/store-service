package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func main() {
	var history []Message
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Chat with Llama (type 'quit' to exit)")
	fmt.Println("--------------------------------------")

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		userInput := scanner.Text()
		if userInput == "quit" {
			break
		}

		history = append(history, Message{
			Role:    "user",
			Content: userInput,
		})

		req := Request{
			Model:    "llama3.2-vision",
			Messages: history,
			Stream:   true,
		}
		if err := streamResponse(req, &history); err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}
	}
}

func streamResponse(req Request, history *[]Message) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Print("\nAssistant: ")

	decoder := json.NewDecoder(resp.Body)
	var assistantMessage Message
	assistantMessage.Role = "assistant"

	for {
		var streamResp StreamResponse
		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding response: %v", err)
		}

		fmt.Print(streamResp.Message.Content)

		assistantMessage.Content += streamResp.Message.Content

		if streamResp.Done {
			break
		}
	}

	*history = append(*history, assistantMessage)
	fmt.Println()

	return nil
}
