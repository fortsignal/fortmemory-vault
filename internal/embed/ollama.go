package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Ollama calls local Ollama /api/embeddings.
type Ollama struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewOllama constructs a client.
func NewOllama(baseURL, model string) *Ollama {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &Ollama{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (o *Ollama) Name() string { return "ollama:" + o.Model }

type embReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embRes struct {
	Embedding []float64 `json:"embedding"`
}

// Embed returns a float32 vector for text.
func (o *Ollama) Embed(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}
	// Cap payload for local models
	if len(text) > 8000 {
		text = text[:8000]
	}
	body, _ := json.Marshal(embReq{Model: o.Model, Prompt: text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := o.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var out embRes
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if len(out.Embedding) == 0 {
		return nil, fmt.Errorf("ollama: empty embedding")
	}
	v := make([]float32, len(out.Embedding))
	for i, f := range out.Embedding {
		v[i] = float32(f)
	}
	return v, nil
}

// Available probes whether Ollama responds (best-effort).
func (o *Ollama) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.BaseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	res, err := o.HTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// None disables embeddings.
type None struct{}

func (None) Name() string { return "none" }
func (None) Embed(context.Context, string) ([]float32, error) {
	return nil, fmt.Errorf("embeddings disabled")
}
