package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
)

type OpenAICompatHandler struct {
	peers *mesh.PeerClient
}

func NewOpenAICompatHandler(peers *mesh.PeerClient) *OpenAICompatHandler {
	return &OpenAICompatHandler{peers: peers}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func (h *OpenAICompatHandler) Enabled() bool {
	value := strings.TrimSpace(os.Getenv("N07_OPENAI_COMPAT_ENABLED"))
	if value == "" {
		return true
	}
	return strings.EqualFold(value, "true") || value == "1"
}

func (h *OpenAICompatHandler) authorize(r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("N07_APP_TOKEN"))
	if expected == "" {
		return false
	}
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	return len(value) > 7 && strings.EqualFold(value[:7], "Bearer ") && strings.TrimSpace(value[7:]) == expected
}

func (h *OpenAICompatHandler) ServeModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": map[string]any{"message": "GET required", "type": "invalid_request_error"}})
		return
	}
	if !h.Enabled() {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{"message": "OpenAI compatibility is disabled", "type": "not_found"}})
		return
	}
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]any{"message": "invalid API credentials", "type": "authentication_error"}})
		return
	}
	if h.peers == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": map[string]any{"message": "Mesh peer client unavailable", "type": "server_error"}})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	description, err := h.peers.Discover(ctx, "N02")
	if err != nil || !meshExecutable(description, openAICapability()) {
		writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": []any{}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data": []map[string]any{{
			"id": "soul-auto", "object": "model", "created": time.Now().Unix(), "owned_by": "SOUL-N02",
		}},
	})
}

func requestCorrelationID(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Request-ID")); value != "" && len(value) <= 200 {
		return value
	}
	return fmt.Sprintf("openai-%d", time.Now().UnixNano())
}

func (h *OpenAICompatHandler) ServeChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": map[string]any{"message": "POST required", "type": "invalid_request_error"}})
		return
	}
	if !h.Enabled() {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{"message": "OpenAI compatibility is disabled", "type": "not_found"}})
		return
	}
	if !h.authorize(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]any{"message": "invalid API credentials", "type": "authentication_error"}})
		return
	}

	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "invalid JSON request", "type": "invalid_request_error"}})
		return
	}
	if strings.TrimSpace(req.Model) != "" && strings.TrimSpace(req.Model) != "soul-auto" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "only model soul-auto is exposed by N07", "type": "invalid_request_error", "param": "model"}})
		return
	}
	if len(req.Messages) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": "messages is required", "type": "invalid_request_error", "param": "messages"}})
		return
	}
	prompt, err := normalizePrompt(req.Messages)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"message": err.Error(), "type": "invalid_request_error", "param": "messages"}})
		return
	}
	if h.peers == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": map[string]any{"message": "Mesh peer client unavailable", "type": "server_error"}})
		return
	}

	correlation := requestCorrelationID(r)
	result, owner, callErr := h.peers.CallBestDynamic(r.Context(), openAICapability(), map[string]any{
		"text": prompt, "source": "openai-compat", "model": "soul-auto",
	}, correlation)
	if callErr != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": callErr.Error(), "type": "upstream_error", "code": "SOUL_MESH_INFERENCE_FAILED"}})
		return
	}

	textValue, err := extractGeneratedText(result)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": map[string]any{"message": err.Error(), "type": "upstream_error", "code": "SOUL_MESH_INVALID_INFERENCE_RESPONSE"}})
		return
	}

	id := fmt.Sprintf("chatcmpl-soul-%d", time.Now().UnixNano())
	created := time.Now().Unix()
	if req.Stream {
		serveSingleChunkStream(w, id, created, "soul-auto", textValue, owner)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"object":  "chat.completion",
		"created": created,
		"model":   "soul-auto",
		"choices": []map[string]any{{
			"index":         0,
			"message":       map[string]string{"role": "assistant", "content": textValue},
			"finish_reason": "stop",
		}},
		"metadata": map[string]any{"n07_owner": owner, "correlation_id": correlation},
	})
}

func normalizePrompt(messages []chatMessage) (string, error) {
	var b strings.Builder
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		if role == "" {
			return "", errors.New("each message requires a role")
		}
		content, err := contentText(message.Content)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(content) == "" {
			continue
		}
		b.WriteString("[")
		b.WriteString(role)
		b.WriteString("]\n")
		b.WriteString(content)
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", errors.New("messages contain no textual content")
	}
	return strings.TrimSpace(b.String()), nil
}

func contentText(content any) (string, error) {
	switch value := content.(type) {
	case string:
		return value, nil
	case []any:
		var parts []string
		for _, item := range value {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if textValue, ok := obj["text"].(string); ok {
				parts = append(parts, textValue)
			}
		}
		return strings.Join(parts, "\n"), nil
	case nil:
		return "", nil
	default:
		return "", errors.New("message content must be text or a text-part array")
	}
}

func extractGeneratedText(result map[string]any) (string, error) {
	for _, key := range []string{"text", "response", "content"} {
		if value, ok := result[key].(string); ok && strings.TrimSpace(value) != "" {
			return value, nil
		}
	}
	return "", errors.New("N02 response did not expose textual output")
}

func openAICapability() string {
	value := strings.TrimSpace(os.Getenv("N07_OPENAI_CAPABILITY"))
	if value == "" {
		return "ai.generate"
	}
	return value
}

func meshExecutable(description map[string]any, capability string) bool {
	raw, ok := description["executableCapabilities"]
	if !ok {
		if nested, nestedOK := description["payload"].(map[string]any); nestedOK {
			raw = nested["executableCapabilities"]
		}
	}
	switch values := raw.(type) {
	case []any:
		for _, item := range values {
			if value, ok := item.(string); ok && strings.SplitN(value, "@", 2)[0] == capability {
				return true
			}
		}
	case []string:
		for _, value := range values {
			if strings.SplitN(value, "@", 2)[0] == capability {
				return true
			}
		}
	}
	return false
}

func serveSingleChunkStream(w http.ResponseWriter, id string, created int64, modelID, textValue, owner string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)
	writeChunk := func(value any) {
		data, _ := json.Marshal(value)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
		if flusher != nil {
			flusher.Flush()
		}
	}
	writeChunk(map[string]any{
		"id": id, "object": "chat.completion.chunk", "created": created, "model": modelID,
		"choices":            []map[string]any{{"index": 0, "delta": map[string]string{"role": "assistant", "content": textValue}, "finish_reason": nil}},
		"system_fingerprint": owner,
	})
	writeChunk(map[string]any{
		"id": id, "object": "chat.completion.chunk", "created": created, "model": modelID,
		"choices": []map[string]any{{"index": 0, "delta": map[string]string{}, "finish_reason": "stop"}},
	})
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}
