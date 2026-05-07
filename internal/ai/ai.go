package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/evanyang0612/git-pr/internal/config"
)

// PRContent is the structured output from the AI
type PRContent struct {
	Title       string
	Description string
}

// Provider is the interface all AI backends implement
type Provider interface {
	Generate(prompt string) (*PRContent, error)
	Name() string
}

// New returns the right provider based on config
func New(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "anthropic":
		return &anthropicProvider{apiKey: cfg.APIKey, model: cfg.Model}, nil
	case "openai":
		return &openaiProvider{apiKey: cfg.APIKey, model: cfg.Model}, nil
	case "gemini":
		return &geminiProvider{apiKey: cfg.APIKey, model: cfg.Model}, nil
	case "ollama":
		return &ollamaProvider{baseURL: cfg.OllamaURL, model: cfg.Model}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}

// ── shared helpers ────────────────────────────────────────

func parseJSON(raw string) (*PRContent, error) {
	// Strip markdown code fences if present
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "{"); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "}"); idx >= 0 && idx < len(raw)-1 {
		raw = raw[:idx+1]
	}

	var result struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parsing AI response: %w\nraw: %s", err, raw)
	}
	if result.Title == "" {
		return nil, fmt.Errorf("AI returned empty title")
	}

	// Unescape \n into real newlines
	result.Description = strings.ReplaceAll(result.Description, `\n`, "\n")

	return &PRContent{
		Title:       result.Title,
		Description: result.Description,
	}, nil
}

func postJSON(url string, headers map[string]string, body any) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// ── Anthropic ─────────────────────────────────────────────

type anthropicProvider struct {
	apiKey string
	model  string
}

func (p *anthropicProvider) Name() string { return "Anthropic " + p.model }

func (p *anthropicProvider) Generate(prompt string) (*PRContent, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": 2048,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	}
	resp, err := postJSON("https://api.anthropic.com/v1/messages", map[string]string{
		"x-api-key":         p.apiKey,
		"anthropic-version": "2023-06-01",
	}, body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from Anthropic")
	}
	return parseJSON(result.Content[0].Text)
}

// ── OpenAI ────────────────────────────────────────────────

type openaiProvider struct {
	apiKey string
	model  string
}

func (p *openaiProvider) Name() string { return "OpenAI " + p.model }

func (p *openaiProvider) Generate(prompt string) (*PRContent, error) {
	body := map[string]any{
		"model":      p.model,
		"max_tokens": 2048,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	}
	resp, err := postJSON("https://api.openai.com/v1/chat/completions", map[string]string{
		"Authorization": "Bearer " + p.apiKey,
	}, body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}
	return parseJSON(result.Choices[0].Message.Content)
}

// ── Gemini ────────────────────────────────────────────────

type geminiProvider struct {
	apiKey string
	model  string
}

func (p *geminiProvider) Name() string { return "Gemini " + p.model }

func (p *geminiProvider) Generate(prompt string) (*PRContent, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)
	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	resp, err := postJSON(url, nil, body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}
	return parseJSON(result.Candidates[0].Content.Parts[0].Text)
}

// ── Ollama (local) ────────────────────────────────────────

type ollamaProvider struct {
	baseURL string
	model   string
}

func (p *ollamaProvider) Name() string { return "Ollama " + p.model }

func (p *ollamaProvider) Generate(prompt string) (*PRContent, error) {
	body := map[string]any{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}
	resp, err := postJSON(p.baseURL+"/api/generate", nil, body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return parseJSON(result.Response)
}
