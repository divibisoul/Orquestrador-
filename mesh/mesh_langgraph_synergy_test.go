package mesh

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sort"
	"testing"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const synergyInput = "SOUL neocortex preserves correlation while routing a neural signal through authorized processing and an agent decision."

type localModelWorker struct {
	cmd *exec.Cmd
	in  *bufio.Writer
	out *bufio.Reader
}

type modelWorkerRequest struct {
	Mode          string `json:"mode"`
	TraceID       string `json:"trace_id"`
	CorrelationID string `json:"correlation_id"`
	InputText     string `json:"input_text"`
	AuthorizedBy  string `json:"authorized_by,omitempty"`
}

type modelWorkerResponse struct {
	Ready              bool      `json:"ready"`
	Error              string    `json:"error,omitempty"`
	TraceID            string    `json:"trace_id"`
	CorrelationID      string    `json:"correlation_id"`
	ModelID            string    `json:"model_id,omitempty"`
	ModelRevisionSHA   string    `json:"model_revision_sha"`
	WeightsSHA256      string    `json:"weights_sha256"`
	OutputSHA256       string    `json:"output_sha256"`
	OutputShape        []int     `json:"output_shape"`
	OutputPreview      []float64 `json:"output_preview"`
	InferenceLatencyMS float64   `json:"inference_latency_ms"`
	Task               string    `json:"task"`
	ModelType          string    `json:"model_type"`
	Tokenizer          string    `json:"tokenizer"`
	TorchVersion       string    `json:"torch_version"`
	TransformersVersion string   `json:"transformers_version"`
	Backend            string    `json:"backend"`
	CheckpointWritten  bool      `json:"checkpoint_written"`
	Checkpointer       string    `json:"checkpointer"`
	GraphNodes         []string  `json:"graph_nodes"`
	LangGraphVersion   string    `json:"langgraph_version,omitempty"`
}

type synergySample struct {
	Index               int       `json:"index"`
	TraceID             string    `json:"trace_id"`
	CorrelationID       string    `json:"correlation_id"`
	LatencyMS           float64   `json:"latency_ms"`
	InferenceLatencyMS  float64   `json:"inference_latency_ms"`
	OutputSHA256        string    `json:"output_sha256"`
	OutputPreview       []float64 `json:"output_preview"`
	ModelRevisionSHA    string    `json:"model_revision_sha"`
	WeightsSHA256       string    `json:"weights_sha256"`
	Task                string    `json:"task"`
	Tokenizer           string    `json:"tokenizer"`
	TorchVersion        string    `json:"torch_version"`
	TransformersVersion string    `json:"transformers_version"`
	Backend             string    `json:"backend"`
	CheckpointWritten   bool      `json:"checkpoint_written"`
	GraphNodes          []string  `json:"graph_nodes"`
}

func startLocalModelWorker(t *testing.T) *localModelWorker {
	t.Helper()
	cmd := exec.Command("python3", "../.github/scripts/mesh_langgraph_worker.py")
	stdin, err := cmd.StdinPipe()
	if err != nil { t.Fatal(err) }
	stdout, err := cmd.StdoutPipe()
	if err != nil { t.Fatal(err) }
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil { t.Fatalf("start local Transformers/LangGraph worker: %v", err) }
	worker := &localModelWorker{cmd: cmd, in: bufio.NewWriter(stdin), out: bufio.NewReader(stdout)}
	line, err := worker.out.ReadBytes('\n')
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("worker did not become ready: %v", err)
	}
	var ready modelWorkerResponse
	if err := json.Unmarshal(bytes.TrimSpace(line), &ready); err != nil || !ready.Ready {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("invalid worker readiness response: raw=%s err=%v", line, err)
	}
	if ready.ModelRevisionSHA != "12040accade4e8a0f71eabdb258fecc2e7e948be" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("unexpected pinned model revision: %s", ready.ModelRevisionSHA)
	}
	if ready.LangGraphVersion == "" || ready.Checkpointer != "MemorySaver" {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		t.Fatalf("LangGraph/checkpointer runtime is not ready: %+v", ready)
	}
	t.Logf("MODEL_WORKER_READY revision=%s weights_sha256=%s backend=%s torch=%s transformers=%s langgraph=%s checkpointer=%s",
		ready.ModelRevisionSHA, ready.WeightsSHA256, ready.Backend, ready.TorchVersion, ready.TransformersVersion, ready.LangGraphVersion, ready.Checkpointer)
	t.Cleanup(func() {
		_, _ = worker.in.WriteString("{\"mode\":\"shutdown\"}\n")
		_ = worker.in.Flush()
		_ = cmd.Wait()
	})
	return worker
}

func (w *localModelWorker) invoke(request modelWorkerRequest) (modelWorkerResponse, error) {
	raw, err := json.Marshal(request)
	if err != nil { return modelWorkerResponse{}, err }
	if _, err := w.in.Write(append(raw, '\n')); err != nil { return modelWorkerResponse{}, err }
	if err := w.in.Flush(); err != nil { return modelWorkerResponse{}, err }
	line, err := w.out.ReadBytes('\n')
	if err != nil { return modelWorkerResponse{}, fmt.Errorf("read local inference worker: %w", err) }
	var response modelWorkerResponse
	if err := json.Unmarshal(bytes.TrimSpace(line), &response); err != nil { return response, err }
	if response.Error != "" { return response, fmt.Errorf("local model worker: %s", response.Error) }
	return response, nil
}

func registerMeshModelCapability(t *testing.T, h *HTTPGateway, worker *localModelWorker, operation, mode string) {
	t.Helper()
	err := h.Engine.Register(operation, func(ctx context.Context, message protocol.Message) (protocol.Result, error) {
		if message.Source != "N01" || message.Target != "N07" {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Status: "rejected", Error: "MESH_SOURCE_OR_TARGET_INVALID"}, fmt.Errorf("expected N01->N07, got %s->%s", message.Source, message.Target)
		}
		if message.Metadata["mesh_contract_version"] != protocol.SoulMeshContractVersion ||
			message.Metadata["mesh_source"] != "N01" || message.Metadata["mesh_target"] != "N07" {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Status: "rejected", Error: "MESH_ENVELOPE_METADATA_INVALID"}, fmt.Errorf("canonical Mesh metadata missing")
		}
		request := modelWorkerRequest{
			Mode: mode, TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			InputText: message.Metadata["input_text"], AuthorizedBy: "N07.Mesh",
		}
		result, err := worker.invoke(request)
		if err != nil { return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Status: "error", Error: err.Error()}, err }
		if result.TraceID != message.TraceID || result.CorrelationID != message.CorrelationID {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Status: "error", Error: "LOCAL_MODEL_TRACE_MISMATCH"}, fmt.Errorf("worker trace/correlation mismatch")
		}
		if mode == "graph" && (!result.CheckpointWritten || result.Checkpointer != "MemorySaver" || len(result.GraphNodes) != 2) {
			return protocol.Result{TraceID: message.TraceID, CorrelationID: message.CorrelationID, Status: "error", Error: "LANGGRAPH_CHECKPOINT_NOT_PROVEN"}, fmt.Errorf("LangGraph checkpoint or graph nodes missing")
		}
		preview, _ := json.Marshal(result.OutputPreview)
		shape, _ := json.Marshal(result.OutputShape)
		nodes, _ := json.Marshal(result.GraphNodes)
		metadata := map[string]string{
			"trace_id": message.TraceID, "correlation_id": message.CorrelationID,
			"model_id": "distilbert/distilbert-base-uncased",
			"model_revision_sha": result.ModelRevisionSHA,
			"weights_sha256": result.WeightsSHA256,
			"output_sha256": result.OutputSHA256,
			"output_shape_json": string(shape), "output_preview_json": string(preview),
			"inference_latency_ms": fmt.Sprintf("%.3f", result.InferenceLatencyMS),
			"task": result.Task, "model_type": result.ModelType, "tokenizer": result.Tokenizer,
			"torch_version": result.TorchVersion, "transformers_version": result.TransformersVersion,
			"backend": result.Backend, "checkpointer": result.Checkpointer,
			"checkpoint_written": fmt.Sprintf("%t", result.CheckpointWritten),
			"graph_nodes_json": string(nodes), "execution_mode": mode,
		}
		return protocol.Result{
			TraceID: message.TraceID, CorrelationID: message.CorrelationID,
			Source: "N07.local-model-adapter", Target: message.Source, Status: "ok", Metadata: metadata,
		}, nil
	})
	if err != nil { t.Fatalf("register Mesh model operation %s: %v", operation, err) }
}

func requestMeshModel(t *testing.T, server *httptest.Server, capability, mode string, index int) synergySample {
	t.Helper()
	traceID := protocol.NewTraceID()
	correlationID := protocol.NewTraceID()
	wire := canonicalRequest("request", capability, correlationID, []float64{1})
	wire["metadata"] = map[string]string{"trace_id": traceID, "input_text": synergyInput}
	body, err := json.Marshal(wire)
	if err != nil { t.Fatal(err) }
	started := time.Now()
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/soul-mesh", bytes.NewReader(body))
	if err != nil { t.Fatal(err) }
	request.Header.Set("Content-Type", "application/json")
	response, err := server.Client().Do(request)
	if err != nil { t.Fatalf("Mesh request %s: %v", capability, err) }
	defer response.Body.Close()
	var envelope struct {
		CorrelationID string         `json:"correlationId"`
		Source        string         `json:"source"`
		Target        string         `json:"target"`
		Kind          string         `json:"kind"`
		Payload       map[string]any `json:"payload"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil { t.Fatal(err) }
	elapsedMS := float64(time.Since(started).Microseconds()) / 1000
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Mesh request failed: status=%d envelope=%+v", response.StatusCode, envelope)
	}
	if envelope.CorrelationID != correlationID || envelope.Source != "N07" || envelope.Target != "N01" || envelope.Kind != "response" {
		t.Fatalf("response envelope/correlation mismatch: %+v", envelope)
	}
	payload, ok := envelope.Payload["metadata"].(map[string]any)
	if !ok { t.Fatalf("response metadata absent: %#v", envelope.Payload) }
	if payload["trace_id"] != traceID || payload["correlation_id"] != correlationID {
		t.Fatalf("trace identity lost on %s: wanted trace=%s correlation=%s got=%#v", mode, traceID, correlationID, payload)
	}
	if payload["execution_mode"] != mode {
		t.Fatalf("mode mismatch: expected %s got %#v", mode, payload["execution_mode"])
	}
	if mode == "graph" && (payload["checkpoint_written"] != "true" || payload["checkpointer"] != "MemorySaver") {
		t.Fatalf("LangGraph checkpointer not proven: %#v", payload)
	}
	preview := []float64{}
	_ = json.Unmarshal([]byte(fmt.Sprint(payload["output_preview_json"])), &preview)
	shape := []int{}
	_ = json.Unmarshal([]byte(fmt.Sprint(payload["output_shape_json"])), &shape)
	nodes := []string{}
	_ = json.Unmarshal([]byte(fmt.Sprint(payload["graph_nodes_json"])), &nodes)
	inferenceLatency := 0.0
	_, _ = fmt.Sscanf(fmt.Sprint(payload["inference_latency_ms"]), "%f", &inferenceLatency)
	return synergySample{
		Index: index, TraceID: traceID, CorrelationID: correlationID, LatencyMS: mathRound3(elapsedMS),
		InferenceLatencyMS: inferenceLatency, OutputSHA256: fmt.Sprint(payload["output_sha256"]),
		OutputPreview: preview, ModelRevisionSHA: fmt.Sprint(payload["model_revision_sha"]),
		WeightsSHA256: fmt.Sprint(payload["weights_sha256"]), Task: fmt.Sprint(payload["task"]),
		Tokenizer: fmt.Sprint(payload["tokenizer"]), TorchVersion: fmt.Sprint(payload["torch_version"]),
		TransformersVersion: fmt.Sprint(payload["transformers_version"]), Backend: fmt.Sprint(payload["backend"]),
		CheckpointWritten: payload["checkpoint_written"] == "true", GraphNodes: nodes,
	}
}

func mathRound3(value float64) float64 { return float64(int(value*1000+0.5)) / 1000 }

func medianMS(values []float64) float64 {
	copyOfValues := append([]float64(nil), values...)
	sort.Float64s(copyOfValues)
	if len(copyOfValues)%2 == 1 { return copyOfValues[len(copyOfValues)/2] }
	return (copyOfValues[len(copyOfValues)/2-1] + copyOfValues[len(copyOfValues)/2]) / 2
}

func TestMeshLangGraphTransformersSynergy10x10(t *testing.T) {
	worker := startLocalModelWorker(t)
	h := newTestGateway(t)
	h.Secret = ""
	h.AllowUnauthenticatedLocal = true
	const directOperation = "soul.model.local.direct@1.0.0"
	const integratedOperation = "soul.model.langgraph.infer@1.0.0"
	registerMeshModelCapability(t, h, worker, directOperation, "direct")
	registerMeshModelCapability(t, h, worker, integratedOperation, "graph")
	server := httptest.NewServer(h)
	defer server.Close()

	// Two warm-ups on each path are excluded from the twenty reported samples.
	for i := 0; i < 2; i++ {
		_ = requestMeshModel(t, server, directOperation, "direct", -i-1)
		_ = requestMeshModel(t, server, integratedOperation, "graph", -i-1)
	}

	baseline := make([]synergySample, 0, 10)
	integrated := make([]synergySample, 0, 10)
	for i := 0; i < 10; i++ {
		baseline = append(baseline, requestMeshModel(t, server, directOperation, "direct", i+1))
	}
	for i := 0; i < 10; i++ {
		integrated = append(integrated, requestMeshModel(t, server, integratedOperation, "graph", i+1))
	}
	baselineLatencies := make([]float64, 0, 10)
	integratedLatencies := make([]float64, 0, 10)
	equivalent := true
	for _, sample := range baseline { baselineLatencies = append(baselineLatencies, sample.LatencyMS) }
	for i, sample := range integrated {
		integratedLatencies = append(integratedLatencies, sample.LatencyMS)
		if i >= len(baseline) || baseline[i].OutputSHA256 != sample.OutputSHA256 ||
			fmt.Sprint(baseline[i].OutputPreview) != fmt.Sprint(sample.OutputPreview) ||
			baseline[i].ModelRevisionSHA != sample.ModelRevisionSHA ||
			baseline[i].WeightsSHA256 != sample.WeightsSHA256 {
			equivalent = false
		}
		if !sample.CheckpointWritten || len(sample.GraphNodes) != 2 {
			t.Fatalf("integrated sample %d did not prove LangGraph checkpoint: %+v", i+1, sample)
		}
	}
	for i, sample := range baseline {
		if sample.ModelRevisionSHA == "" || sample.WeightsSHA256 == "" || sample.OutputSHA256 == "" {
			t.Fatalf("baseline sample %d lacks model/output provenance: %+v", i+1, sample)
		}
		if i > 0 && (baseline[i-1].OutputSHA256 != sample.OutputSHA256 ||
			fmt.Sprint(baseline[i-1].OutputPreview) != fmt.Sprint(sample.OutputPreview)) {
			equivalent = false
		}
	}

	p50Baseline := medianMS(baselineLatencies)
	p50Integrated := medianMS(integratedLatencies)
	ratio := p50Integrated / p50Baseline
	verdict := "UNPROVEN"
	if equivalent && p50Baseline > 0 && ratio <= 1.5 { verdict = "PROVEN" }
	report := map[string]any{
		"stage": "SOUL_MESH_SYNERGY_10_PLUS_10",
		"status": verdict,
		"mesh_transport": "HTTPGateway via test-local Mesh endpoint; source N01 target N07",
		"model_id": "distilbert/distilbert-base-uncased",
		"model_revision_sha": baseline[0].ModelRevisionSHA,
		"weights_sha256": baseline[0].WeightsSHA256,
		"backend": baseline[0].Backend,
		"task": baseline[0].Task,
		"tokenizer": baseline[0].Tokenizer,
		"torch_version": baseline[0].TorchVersion,
		"transformers_version": baseline[0].TransformersVersion,
		"checkpointer": "MemorySaver",
		"warmups_each_path_excluded": 2,
		"baseline_samples": baseline,
		"integrated_samples": integrated,
		"p50_baseline_ms": mathRound3(p50Baseline),
		"p50_integrated_ms": mathRound3(p50Integrated),
		"p50_ratio": mathRound3(ratio),
		"outputs_equivalent": equivalent,
		"criterion_ratio_max": 1.5,
	}

	raw, err := json.Marshal(report)
	if err != nil { t.Fatal(err) }
	fmt.Printf("SOUL_MESH_SYNERGY_10_PLUS_10=%s\n", raw)
	if verdict != "PROVEN" {
		t.Fatalf("synergy criterion not met: p50_integrated/p50_baseline=%.3f, outputs_equivalent=%t", ratio, equivalent)
	}
}

