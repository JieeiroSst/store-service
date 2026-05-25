package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type openAIService struct {
	cfg           Config
	httpClient    *http.Client
	candidateRepo candidate.Repository
	logger        *zap.Logger
}

func NewOpenAIService(cfg Config, candidateRepo candidate.Repository, logger *zap.Logger) port.AIService {
	return &openAIService{
		cfg:           cfg,
		httpClient:    &http.Client{Timeout: 60 * time.Second},
		candidateRepo: candidateRepo,
		logger:        logger,
	}
}


func (s *openAIService) ScoreMatch(ctx context.Context, input port.ScoreMatchInput) (shared.AIScore, error) {
	system := `You are a senior technical recruiter AI. Score the candidate-job match from 0-100.
Return ONLY valid JSON: {"score":85,"confidence":0.9,"breakdown":{"skills":90,"experience":80,"culture":85}}`

	prompt := fmt.Sprintf("Job ID: %s\nCandidate ID: %s\nScore this match.", input.JobID, input.CandidateID)

	raw, err := s.chat(ctx, system, prompt)
	if err != nil {
		return shared.AIScore{}, err
	}

	var result struct {
		Score      float64            `json:"score"`
		Confidence float64            `json:"confidence"`
		Breakdown  map[string]float64 `json:"breakdown"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return shared.AIScore{}, fmt.Errorf("ai: parse score response: %w", err)
	}

	return shared.AIScore{
		Score:      result.Score,
		Confidence: result.Confidence,
		Breakdown:  result.Breakdown,
		ScoredAt:   time.Now(),
		ModelID:    s.cfg.Model,
	}, nil
}

func (s *openAIService) ParseResume(ctx context.Context, resumeURL string) (*port.ParseResumeOutput, error) {
	system := `You are a resume parser. Extract structured information.
Return ONLY valid JSON:
{"skills":["Go","PostgreSQL"],"years_experience":5,"experience_level":"senior","suggested_title":"Backend Engineer"}`

	raw, err := s.chat(ctx, system, "Parse this resume: "+resumeURL)
	if err != nil {
		return nil, err
	}

	var result struct {
		Skills          []string `json:"skills"`
		YearsExperience int      `json:"years_experience"`
		ExperienceLevel string   `json:"experience_level"`
		SuggestedTitle  string   `json:"suggested_title"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("ai: parse resume response: %w", err)
	}

	embedding, err := s.embed(ctx, resumeURL)
	if err != nil {
		s.logger.Warn("embedding failed, continuing without", zap.Error(err))
		embedding = nil
	}

	return &port.ParseResumeOutput{
		Skills:          result.Skills,
		YearsExperience: result.YearsExperience,
		ExperienceLevel: candidate.ExperienceLevel(result.ExperienceLevel),
		SuggestedTitle:  result.SuggestedTitle,
		Embedding:       embedding,
	}, nil
}

func (s *openAIService) RecommendCandidates(ctx context.Context, jobID uuid.UUID, limit int) ([]*candidate.Candidate, error) {
	return s.candidateRepo.FindSimilar(ctx, nil, limit)
}


func (s *openAIService) GenerateJobDescription(ctx context.Context, title, requirements string) (string, error) {
	system := `You are an expert HR copywriter. Write compelling, inclusive job descriptions.
Be specific, avoid jargon, and highlight growth opportunities.`
	prompt := fmt.Sprintf("Job title: %s\nRequirements: %s\n\nWrite a full job description.", title, requirements)
	return s.chat(ctx, system, prompt)
}

func (s *openAIService) chat(ctx context.Context, system, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model": s.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"temperature":     0.2,
		"response_format": map[string]string{"type": "json_object"},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.BaseURL+"/chat/completions", bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ai: OpenAI error %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ai: decode error: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("ai: no choices returned")
	}
	return result.Choices[0].Message.Content, nil
}

func (s *openAIService) embed(ctx context.Context, text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.BaseURL+"/embeddings", bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("ai: no embedding returned")
	}
	return result.Data[0].Embedding, nil
}
