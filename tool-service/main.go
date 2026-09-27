package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/tool-service/internal/handler"
	"github.com/JIeeiroSst/tool-service/internal/learning"
	"github.com/JIeeiroSst/tool-service/internal/ollama"
	"github.com/JIeeiroSst/tool-service/internal/suite"
)

//go:embed web
var webFS embed.FS

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Next()
	}
}

func loadInfra(dataDir string) handler.Infra {
	in := handler.Infra{DataDir: dataDir, NoSandbox: os.Getenv("CHROME_NO_SANDBOX") == "true"}
	if v := os.Getenv("DB_CONNECTIONS"); v != "" {
		if err := json.Unmarshal([]byte(v), &in.DBs); err != nil {
			log.Fatal("DB_CONNECTIONS is not valid JSON: ", err)
		}
	}
	if v := os.Getenv("RABBITMQ_CONNECTIONS"); v != "" {
		if err := json.Unmarshal([]byte(v), &in.Rabbit); err != nil {
			log.Fatal("RABBITMQ_CONNECTIONS is not valid JSON: ", err)
		}
	}
	return in
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	model := env("OLLAMA_MODEL", "llama3.2")
	embedModel := env("OLLAMA_EMBED_MODEL", "nomic-embed-text")

	ai := ollama.NewClient(env("OLLAMA_URL", "http://localhost:11434"), model, 5*time.Minute)
	mem, err := learning.Open(env("LEARNING_DATA_DIR", "./data"))
	if err != nil {
		log.Fatal("open learning store: ", err)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if err := ai.Pull(ctx, embedModel); err != nil {
			log.Printf("embedding model %q not available, recall falls back to recency: %v", embedModel, err)
		}
	}()

	suites, err := suite.Open(env("LEARNING_DATA_DIR", "./data"))
	if err != nil {
		log.Fatal("open suites: ", err)
	}
	runs, _ := strconv.Atoi(env("CONSISTENCY_RUNS", "3"))
	h := handler.New(ai, mem, handler.LearnConfig{
		BaseModel:  model,
		EmbedModel: embedModel,
		AgentName:  env("QC_AGENT_NAME", "qc-agent"),
		FewShot:    3,

		ConsistencyRuns: runs,
	}, strings.Split(os.Getenv("TARGET_ALLOWLIST"), ","))

	h.SetSuites(suites)
	h.SetInfra(loadInfra(env("LEARNING_DATA_DIR", "./data")))
	h.SetAPIToken(os.Getenv("API_TOKEN"))
	if os.Getenv("API_TOKEN") == "" {
		log.Print("WARNING: API_TOKEN is not set, /api/v1 is open to anyone who can reach this service")
	}

	r := gin.Default()
	r.Use(securityHeaders())
	serve := func(file, ctype string) gin.HandlerFunc {
		b, err := webFS.ReadFile("web/" + file)
		if err != nil {
			log.Fatal(err)
		}
		return func(c *gin.Context) { c.Data(http.StatusOK, ctype, b) }
	}
	r.GET("/", serve("index.html", "text/html; charset=utf-8"))
	r.GET("/assets/app.js", serve("app.js", "text/javascript; charset=utf-8"))
	r.GET("/assets/app.css", serve("app.css", "text/css; charset=utf-8"))
	h.Register(r)

	if d, err := time.ParseDuration(env("AUTO_LEARN_INTERVAL", "1h")); err == nil && d > 0 {
		evolveEvery, _ := strconv.Atoi(env("EVOLVE_EVERY_APPROVED", "20"))
		go h.AutoLearn(context.Background(), d, 5, evolveEvery)
	}

	log.Fatal(r.Run(":" + env("HTTP_PORT", "8080")))
}
