package superagi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type GeminiProvider struct {
	Model string
	APIKey string
	HTTP  *http.Client
}

func NewGeminiProviderFromEnv(client *http.Client) *GeminiProvider {
	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if model == "" {
		model = "gemini-2.5-flash"
	}
	key := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &GeminiProvider{Model: model, APIKey: key, HTTP: client}
}

func (g *GeminiProvider) Generate(ctx context.Context, prompt string) (string, error) {
	if g == nil { return "", errors.New("nil Gemini provider") }
	if ctx == nil { return "", errors.New("nil context") }
	if err := ctx.Err(); err != nil { return "", err }
	prompt = strings.TrimSpace(prompt)
	if prompt == "" { return "", errors.New("prompt required") }
	if strings.TrimSpace(g.APIKey) == "" { return "", errors.New("Gemini provider is not configured") }

	body := map[string]any{"contents": []any{map[string]any{"parts": []any{map[string]string{"text": prompt}}}}}
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct { Text string `json:"text"` } `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := g.post(ctx, "models/"+url.PathEscape(g.Model)+":generateContent", body, &response); err != nil { return "", err }
	for _, c := range response.Candidates {
		for _, p := range c.Content.Parts {
			if text := strings.TrimSpace(p.Text); text != "" { return text, nil }
		}
	}
	return "", errors.New("Gemini returned no text")
}

func (g *GeminiProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	if g == nil { return nil, errors.New("nil Gemini provider") }
	if ctx == nil { return nil, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return nil, err }
	text = strings.TrimSpace(text)
	if text == "" { return nil, errors.New("text required") }
	if strings.TrimSpace(g.APIKey) == "" { return nil, errors.New("Gemini provider is not configured") }

	body := map[string]any{"content": map[string]any{"parts": []any{map[string]string{"text": text}}}}
	var response struct { Embedding struct { Values []float64 `json:"values"` } `json:"embedding"` }
	model := g.Model
	if model == "gemini-2.5-flash" { model = "text-embedding-004" }
	if err := g.post(ctx, "models/"+url.PathEscape(model)+":embedContent", body, &response); err != nil { return nil, err }
	if len(response.Embedding.Values) == 0 { return nil, errors.New("Gemini returned no embedding") }
	return response.Embedding.Values, nil
}

func (g *GeminiProvider) post(ctx context.Context, action string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil { return fmt.Errorf("marshal Gemini request: %w", err) }
	endpoint := "https://generativelanguage.googleapis.com/v1beta/" + strings.TrimPrefix(action, "/")
	u, err := url.Parse(endpoint)
	if err != nil { return fmt.Errorf("Gemini endpoint: %w", err) }
	q := u.Query(); q.Set("key", g.APIKey); u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(payload))
	if err != nil { return fmt.Errorf("create Gemini request: %w", err) }
	req.Header.Set("Content-Type", "application/json")
	client := g.HTTP
	if client == nil { client = &http.Client{Timeout: 30 * time.Second} }
	resp, err := client.Do(req)
	if err != nil { return fmt.Errorf("Gemini request: %w", err) }
	defer resp.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return fmt.Errorf("Gemini API %s: %s", resp.Status, strings.TrimSpace(string(data))) }
	if err := json.Unmarshal(data, out); err != nil { return fmt.Errorf("decode Gemini response: %w", err) }
	return nil
}
