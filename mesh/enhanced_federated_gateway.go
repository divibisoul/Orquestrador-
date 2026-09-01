package mesh

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type EnhancedFederatedGateway struct {
	base *FederatedGateway
}

func NewEnhancedFederatedHTTPGateway(engine *orchestrator.Engine) *EnhancedFederatedGateway {
	return &EnhancedFederatedGateway{base: NewFederatedHTTPGateway(engine)}
}

func (g *EnhancedFederatedGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if g == nil || g.base == nil {
		writeMeshJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "N07 gateway unavailable"})
		return
	}
	if r.Method != http.MethodPost || g.base.engine == nil {
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
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		g.base.ServeHTTP(w, r)
		return
	}
	capability := canonicalCapability(wire)
	if capability != "mesh.supergpu.parallel" && capability != "supergpu.parallel" {
		r.Body = io.NopCloser(strings.NewReader(string(body)))
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
		ID         string
		Capability string
		Payload    map[string]any
		Required   bool
	}
	tasks := make([]task, 0, len(rawTasks))
	seenIDs := map[string]struct{}{}
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
		id = strings.TrimSpace(id)
		if id == "" {
			id = "task-" + jsonNumber(index)
		}
		if _, exists := seenIDs[id]; exists {
			g.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "duplicate task id", "id": id})
			return
		}
		seenIDs[id] = struct{}{}
		stepPayload, _ := item["payload"].(map[string]any)
		if stepPayload == nil {
			stepPayload = map[string]any{}
		}
		required, _ := item["required"].(bool)
		tasks = append(tasks, task{ID: id, Capability: capability, Payload: stepPayload, Required: required})
	}

	type result struct {
		ID            string `json:"id"`
		Capability    string `json:"capability"`
		Owner         string `json:"owner,omitempty"`
		Status        string `json:"status"`
		DurationMs    int64  `json:"duration_ms"`
		Payload       any    `json:"payload,omitempty"`
		Error         string `json:"error,omitempty"`
		CorrelationID string `json:"correlationId"`
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
			remote, owner, callErr := g.base.peers.CallBest(r.Context(), item.Capability, item.Payload, childCorrelation)
			entry := result{ID: item.ID, Capability: item.Capability, Status: "error", DurationMs: time.Since(started).Milliseconds(), CorrelationID: childCorrelation}
			if callErr != nil {
				entry.Error = callErr.Error()
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
		"execution":           "federated-supergpu-parallel",
		"parentCorrelationId": envelope.CorrelationID,
		"taskCount":           len(results),
		"requiredFailure":     requiredFailed,
		"tasks":               results,
	})
}

func jsonNumber(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "0"
	}
	return string(b)
}
