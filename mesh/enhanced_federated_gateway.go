package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxFederatedParallelTasks = 32
const federatedTaskTimeout = 15 * time.Second

type EnhancedFederatedGateway struct{ base *FederatedGateway }

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
	if err := g.base.base.authenticateWire(wire, r, envelope); err != nil {
		g.base.base.respond(w, http.StatusUnauthorized, envelope, "ERROR", map[string]any{"error": "mesh authentication failed"})
		return
	}
	payload := envelope.NestedPayload()
	raw, ok := payload["tasks"].([]any)
	if !ok || len(raw) == 0 {
		g.base.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "supergpu.parallel requires payload.tasks"})
		return
	}
	if len(raw) > maxFederatedParallelTasks {
		g.base.base.respond(w, http.StatusRequestEntityTooLarge, envelope, "ERROR", map[string]any{"error": "task count exceeds configured parallel limit", "limit": maxFederatedParallelTasks})
		return
	}
	type task struct {
		ID         string
		Capability string
		Payload    map[string]any
		Required   bool
		Timeout    time.Duration
	}
	tasks := make([]task, 0, len(raw))
	seen := map[string]struct{}{}
	for i, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			g.base.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "task must be object", "index": i})
			return
		}
		capability, _ := m["capability"].(string)
		capability = strings.TrimSpace(capability)
		if capability == "" {
			g.base.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "task capability is required", "index": i})
			return
		}
		id, _ := m["id"].(string)
		id = strings.TrimSpace(id)
		if id == "" {
			id = fmt.Sprintf("task-%d", i)
		}
		if _, exists := seen[id]; exists {
			g.base.base.respond(w, http.StatusBadRequest, envelope, "ERROR", map[string]any{"error": "duplicate task id", "id": id})
			return
		}
		seen[id] = struct{}{}
		p, _ := m["payload"].(map[string]any)
		if p == nil {
			p = map[string]any{}
		}
		required, _ := m["required"].(bool)
		timeout := federatedTaskTimeout
		if rawTimeout, ok := m["timeout_ms"].(float64); ok && rawTimeout > 0 && rawTimeout < float64(timeout/time.Millisecond) {
			timeout = time.Duration(rawTimeout) * time.Millisecond
		}
		tasks = append(tasks, task{id, capability, p, required, timeout})
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
	sem := make(chan struct{}, maxFederatedParallelTasks)
	ctx := r.Context()
	for i, item := range tasks {
		wg.Add(1)
		go func(index int, current task) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				results[index] = result{ID: current.ID, Capability: current.Capability, Status: "cancelled", Error: ctx.Err().Error(), CorrelationID: envelope.CorrelationID + "/" + current.ID}
				return
			}
			defer func() { <-sem }()
			started := time.Now()
			child := envelope.CorrelationID + "/" + current.ID
			taskCtx, cancel := context.WithTimeout(ctx, current.Timeout)
			defer cancel()
			remote, owner, e := g.base.peers.CallBestDynamic(taskCtx, current.Capability, current.Payload, child)
			entry := result{ID: current.ID, Capability: current.Capability, Status: "error", DurationMs: time.Since(started).Milliseconds(), CorrelationID: child}
			if e != nil {
				entry.Error = e.Error()
				if taskCtx.Err() != nil {
					entry.Status = "cancelled"
					entry.Error = taskCtx.Err().Error()
				}
			} else {
				entry.Status = "ok"
				entry.Owner = owner
				entry.Payload = remote["payload"]
			}
			results[index] = entry
		}(i, item)
	}
	wg.Wait()
	requiredFailed := false
	for i, v := range results {
		if v.Status != "ok" && tasks[i].Required {
			requiredFailed = true
		}
	}
	status := http.StatusOK
	if requiredFailed {
		status = http.StatusBadGateway
	}
	g.base.base.respond(w, status, envelope, "TASK_RESULT", map[string]any{"execution": "federated-supergpu-parallel", "parentCorrelationId": envelope.CorrelationID, "taskCount": len(results), "parallelLimit": maxFederatedParallelTasks, "requiredFailure": requiredFailed, "tasks": results})
}
