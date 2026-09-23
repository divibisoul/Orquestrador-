package octacore

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
    "time"

    "github.com/divibisoul/Orquestrador-/protocol"
)

type recordingControl struct {
    mu sync.Mutex
    events []VagusEnvelope
}

func (r *recordingControl) Publish(_ context.Context, event VagusEnvelope) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.events = append(r.events, event)
    return nil
}

func (r *recordingControl) Status() string { return "TEST_CONTROL_REAL" }

func TestFederatedContextCycleFullContractFlow(t *testing.T) {
    const secret = "octacore-contract-secret-0123456789"

    sara := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        if req.Header.Get("Authorization") != "Bearer sara-test-token" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        var body map[string]any
        if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        switch req.URL.Path {
        case "/v1/audit":
            _ = json.NewEncoder(w).Encode(map[string]any{
                "cycle_id": body["input"],
                "converged": true,
                "trace_hash": "audit-trace-" + req.Header.Get("X-Correlation-ID"),
            })
        case "/v1/cycle":
            contextValue, ok := body["context"].(map[string]any)
            if !ok || contextValue == nil {
                http.Error(w, "missing context", http.StatusBadRequest)
                return
            }
            _ = json.NewEncoder(w).Encode(map[string]any{
                "cycle_id": body["cycle_id"],
                "correlation_id": req.Header.Get("X-Correlation-ID"),
                "converged": true,
                "rollback_performed": false,
                "trace_hash": "cycle-trace-" + req.Header.Get("X-Correlation-ID"),
                "execution_report": map[string]any{"evidence_hash": "evidence-" + req.Header.Get("X-Correlation-ID")},
                "final_state": map[string]any{"pipeline_status": contextValue["pipeline_status"]},
            })
        default:
            http.NotFound(w, req)
        }
    }))
    defer sara.Close()

    makeMeshPeer := func(nucleus string, handler func(map[string]any) map[string]any) *httptest.Server {
        return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
            var envelope map[string]any
            if err := json.NewDecoder(req.Body).Decode(&envelope); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            if envelope["protocol"] != "soul-mesh/1" ||
                envelope["contractVersion"] != protocol.SoulMeshContractVersion ||
                envelope["source"] != protocol.N07 ||
                envelope["target"] != nucleus {
                http.Error(w, "invalid mesh envelope", http.StatusBadRequest)
                return
            }
            correlation, _ := envelope["correlationId"].(string)
            capability, _ := envelope["capability"].(string)
            payload, _ := envelope["payload"].(map[string]any)
            response := protocol.MeshEnvelope{
                Version: protocol.SoulMeshVersion,
                ContractVersion: protocol.SoulMeshContractVersion,
                MessageID: protocol.NewTraceID(),
                Source: nucleus,
                Target: protocol.N07,
                Timestamp: time.Now().UnixMilli(),
                Nonce: protocol.NewTraceID(),
                CorrelationID: correlation,
                Type: "TASK_RESULT",
                Payload: map[string]any{
                    "capability": capability,
                    "payload": handler(payload),
                },
            }
            if err := protocol.SignHMAC(&response, secret); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            wire := map[string]any{
                "protocol": "soul-mesh/1",
                "contractVersion": response.ContractVersion,
                "id": response.MessageID,
                "correlationId": response.CorrelationID,
                "source": response.Source,
                "target": response.Target,
                "kind": "response",
                "capability": capability,
                "payload": response.Payload["payload"],
                "timestamp": response.Timestamp,
                "nonce": response.Nonce,
                "hmac": response.HMAC,
            }
            w.Header().Set("Content-Type", "application/json")
            _ = json.NewEncoder(w).Encode(wire)
        }))
    }

    n04 := makeMeshPeer("N04", func(_ map[string]any) map[string]any {
        return map[string]any{
            "research_snippets": []any{map[string]any{"source": "N04", "text": "contract-research"}},
            "pipeline": "research_ready",
        }
    })
    defer n04.Close()

    n03 := makeMeshPeer("N03", func(_ map[string]any) map[string]any {
        return map[string]any{"perception": map[string]any{"prepared": true, "source": "N03"}}
    })
    defer n03.Close()

    t.Setenv("SARA_SERVICE_URL", sara.URL)
    t.Setenv("SARA_SERVICE_TOKEN", "sara-test-token")
    t.Setenv("SOUL_MESH_N04_URL", n04.URL)
    t.Setenv("SOUL_MESH_N03_URL", n03.URL)
    t.Setenv("SOUL_MESH_HMAC_SECRET", secret)

    control := &recordingControl{}
    processor, err := NewProcessor(DefaultSchedulerConfig(), control)
    if err != nil {
        t.Fatal(err)
    }

    correlation := "g6-g4-g3-g0-contract-001"
    result := processor.ExecuteFederatedContextCycle(context.Background(), FederatedContextInput{
        CorrelationID: correlation,
        Input: "federated contract test",
        ResearchPayload: map[string]any{"query": "real context"},
        PerceptionPayload: map[string]any{"capability": "mesh.describe"},
        AllowResearchSkip: false,
        TTLMS: 10_000,
    })

    if !result.Audit.OK {
        t.Fatalf("G0 audit failed: %#v", result.Audit.Error)
    }
    if !result.Cycle.OK {
        t.Fatalf("G0 cycle failed: %#v", result.Cycle.Error)
    }
    if result.CorrelationID != correlation || result.Cycle.CorrelationID != correlation {
        t.Fatalf("correlation was not preserved: result=%q cycle=%q", result.CorrelationID, result.Cycle.CorrelationID)
    }
    if result.Research == nil || result.Perception == nil {
        t.Fatalf("expected both G4 research and G3 perception outputs: %#v", result)
    }
    control.mu.Lock()
    defer control.mu.Unlock()
    var submit, resultEvents, barrier int
    for _, event := range control.events {
        switch event.Type {
        case "gpu.submit":
            submit++
        case "gpu.result":
            resultEvents++
        case "gpu.barrier":
            barrier++
        }
    }
    if submit < 2 || resultEvents < 2 || barrier != 1 {
        t.Fatalf("unexpected Octacore control events: submit=%d result=%d barrier=%d events=%#v", submit, resultEvents, barrier, control.events)
    }
}
