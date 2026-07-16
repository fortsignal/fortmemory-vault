// Package embed provides optional local embeddings (Ollama HTTP).
package embed

import "context"

// Provider turns text into a vector.
type Provider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Name() string
}
