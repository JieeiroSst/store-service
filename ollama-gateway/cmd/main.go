package main

import (
	"flag"
	"log"

	"ollama-gateway/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "WebSocket server address")
	presetsDir := flag.String("presets", "./presets", "Directory containing preset Q&A files")
	ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama base URL")
	ollamaModel := flag.String("model", "llama3.2", "Ollama model name")
	flag.Parse()

	cfg := server.Config{
		Addr:        *addr,
		PresetsDir:  *presetsDir,
		OllamaURL:   *ollamaURL,
		OllamaModel: *ollamaModel,
	}

	log.Printf("🚀 Starting Ollama Gateway on %s", cfg.Addr)
	log.Printf("📂 Presets directory: %s", cfg.PresetsDir)
	log.Printf("🤖 Ollama: %s (model: %s)", cfg.OllamaURL, cfg.OllamaModel)

	if err := server.Run(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
