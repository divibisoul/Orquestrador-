package mesh

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type EnhancedFederatedGateway struct {
	base *FederatedGateway
}

func NewEnhancedFederatedHTTPGateway(engine *orchestrator.Engine) *EnhancedFederatedGateway {
	return &EnhancedFederatedGateway{base: NewFederatedHTTPGateway(engine)}
}

func (g *EnhancedFederatedGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || g == nil || g.base == nil || g.base.engine == nil {
		g.base.ServeHTTP(w, r)
		return
	}

	body, err := ioReadLimited(r.Body, 1<<20)
	if err != nil {
		writeMeshJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	var wire canonicalWireEnvelope
	if err := json.Unmarshal(body, &wire); err != nil {
		r.Body = ioNopCloserString(body)
		g.base.ServeHTTP(w, r)
		return
	}
	capability := canonicalCapability(wire)
	if capability != "mesh.supergpu.parallel" && capability != "supergpu.parallel" {
		r.Body = ioNopCloserString(body)
		g.base.ServeHTTP(w, r)
		return
	}

	envelope := normalizedMeshEnvelope(wire)
	if err := envelope.Validate(); err != nil {
		writeMeshJSON(w, http.StatusBadRequest, gatewayError(envelope, err.Error()))
		return
	}
	if err := g.base.authenticateWire(wire, r, envelope); err != nil {
		g.base.respond(w, http.StatusUnauthorized, envelope, "ERROR", map[string]any{"error": "mesh authentication failed"})
		return
	}

	payload := envelope.NestedPayload()
	if payload == nil {
		payload = map[string]any{}
	}
	rawTasks, ok := payload["tasks"].([]any)
	if !ok || len(rawTasks) == 0 {
		g.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "supergpu.parallel requires payload.tasks"})
		return
	}

	type task struct {
		ID       string
		Capability string
		Payload  map[string]any
		Required bool
	}
	tasks := make([]task, 0, len(rawTasks))
	for index, raw := range rawTasks {
		item, ok := raw.(map[string]any)
		if !ok {
			g.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "each task must be an object", "index": index})
			return
		}
		capability, _ := item["capability"].(string)
		capability = strings.TrimSpace(capability)
		if capability == "" {
			g.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "task capability is required", "index": index})
			return
		}
		id, _ := item["id"].(string)
		if strings.TrimSpace(id) == "" {
			id = "task-" + jsonNumber(index)
		}
		stepPayload, _ := item["payload"].(map[string]any)
		if stepPayload == nil {
			stepPayload = map[string]any{}
		}
		required, _ := item["required"].(bool)
		tasks = append(tasks, task{ID: id, Capability: capability, Payload: stepPayload, Required: required})
	}

	type result struct {
		ID string `json:"id"`
		Capability string `json:"capability"`
		Owner string `json:"owner,omitempty"`
		Status string `json:"status"`
		DurationMs int64 `json:"duration_ms"`
		Payload any `json:"payload,omitempty"`
		Error string `json:"error,omitempty"`
	}
	results := make([]result, len(tasks))
	var wg sync.WaitGroup
	for index, item := range tasks {
		index, item := index, item
		wg.Add(1)
		go func() {
			defer wg.Done()
			started := time.Now()
			childCorrelation := envelope.CorrelationID + "/" + item.ID
			remote, owner, err := g.base.peers.CallBest(r.Context(), item.Capability, item.Payload, childCorrelation)
			entry := result{ID: item.ID, Capability: item.Capability, Status: "error", DurationMs: time.Since(started).Milliseconds()}
			if err != nil {
				entry.Error = err.Error()
			} else {
				entry.Status = "ok"
				entry.Owner = owner
				entry.Payload = remote["payload"]
			}
			results[index] = entry
		}()
	}
	wg.Wait()

	requiredFailed := false
	for index, entry := range results {
		if entry.Status != "ok" && tasks[index].Required {
			requiredFailed = true
		}
	}
	responseStatus := http.StatusOK
	if requiredFailed {
		responseStatus = http.StatusBadGateway
	}
	g.base.respond(w, responseStatus, envelope, "TASK_RESULT", map[string]any{
		"execution": "federated-supergpu-parallel",
		"parentCorrelationId": envelope.CorrelationID,
		"taskCount": len(results),
		"requiredFailure": requiredFailed,
		"tasks": results,
	})
}

func jsonNumber(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(b)
}

func ioNopCloserString(body []byte) *readCloser {
	return &readCloser{data: body}
}

type readCloser struct { data []byte; done bool }
func (r *readCloser) Read(p []byte) (int, error) {
	if r.done { return 0, errors.New("EOF") }
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 { r.done = true }
	return n, nil
}
func (r *readCloser) Close() error { r.done = true; r.data = nil; return nil }
