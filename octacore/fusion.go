package octacore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

const (
	OpFusionDescribe = "octacore.fusion.describe@1.0.0"
	OpFusionExecute  = "octacore.fusion.execute@1.0.0"
)

type Fusion struct {
	engine *orchestrator.Engine
	peers  *mesh.PeerClient
	gpu    *supergpu.Runtime
	agent  *SuperpowersOctacoreAgent
}

func NewFusion(engine *orchestrator.Engine, peers *mesh.PeerClient, gpu *supergpu.Runtime) (*Fusion, error) {
	if engine == nil {
		return nil, errors.New("Octacore fusion requires N07 orchestrator")
	}
	if peers == nil {
		return nil, errors.New("Octacore fusion requires canonical Mesh peer client")
	}
	if gpu == nil {
		return nil, errors.New("Octacore fusion requires existing N07 SuperGPU runtime")
	}
	return &Fusion{engine: engine, peers: peers, gpu: gpu, agent: NewSuperpowersOctacoreAgent()}, nil
}

func (f *Fusion) Register() error {
	if err := f.engine.Register(OpFusionDescribe, f.describe); err != nil {
		return err
	}
	if err := f.engine.Register(OpFusionExecute, f.execute); err != nil {
		return err
	}
	return nil
}

func (f *Fusion) describe(ctx context.Context, m protocol.Message) (protocol.Result, error) {
	payload := map[string]any{
		"name":         "N07 Octacore Fusion",
		"architecture": "orchestrator+prefrontal-neocortex+TCE-orbital-reasoner+SuperGPU",
		"transport":    "soul-mesh/1",
		"ownership": map[string]string{
			"orchestrator":           "N07",
			"prefrontal":             "N07",
			"orbital_reasoner":       "N07",
			"supergpu_control_plane": "N07",
		},
		"stages": []string{
			"transcendental.estimate@1.0.0",
			"prefrontal.orbital.evaluate@1.0.0",
			"supergpu.execute@1.0.0",
		},
		"fail_closed":    true,
		"hardware_claim": "none",
		"superagi_affinity": []string{
			"mlfg",
			"skill_acquisition",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return protocol.Result{}, err
	}
	return protocol.Result{
		TraceID: m.TraceID, CorrelationID: m.CorrelationID,
		Source: "N07.octacore", Target: m.Source, Status: "ok",
		Metadata: map[string]string{"octacore_fusion_json": string(raw)},
	}, nil
}

func (f *Fusion) execute(ctx context.Context, m protocol.Message) (protocol.Result, error) {
	if ctx == nil {
		return protocol.Result{}, errors.New("context is nil")
	}
	if err := f.agent.Preflight(OpFusionExecute); err != nil {
		return f.fail(m, "SUPERPOWERS_AGENT_PREFLIGHT_FAILED", err)
	}
	workloads := strings.TrimSpace(m.Metadata["workloads_json"])
	candidate := strings.TrimSpace(m.Metadata["candidate_json"])
	if workloads == "" || candidate == "" {
		return f.fail(m, "FUSION_INPUT_REQUIRED", errors.New("workloads_json and candidate_json are required"))
	}

	start := time.Now()

	estimate, err := f.engine.Execute(ctx, "transcendental.estimate@1.0.0", m.Payload, map[string]string{
		"workloads_json": workloads,
		"strategy":       strings.TrimSpace(m.Metadata["strategy"]),
		"correlation_id": m.CorrelationID,
	})
	if err != nil {
		return f.fail(m, "ORBITAL_STAGE_FAILED", err)
	}

	admission, err := f.engine.Execute(ctx, "prefrontal.orbital.evaluate@1.0.0", m.Payload, map[string]string{
		"workloads_json": workloads,
		"candidate_json": candidate,
		"strategy":       strings.TrimSpace(m.Metadata["strategy"]),
		"correlation_id": m.CorrelationID,
	})
	if err != nil {
		return f.fail(m, "PREFRONTAL_STAGE_FAILED", err)
	}

	device := strings.TrimSpace(m.Metadata["device"])
	if device == "" {
		devices := f.gpu.Discover()
		if len(devices) == 0 {
			return f.fail(m, "SUPERGPU_NO_DEVICE", errors.New("no executable SuperGPU device is available"))
		}
		device = devices[0].ID
	}
	operation := strings.TrimSpace(m.Metadata["operation"])
	if operation == "" {
		return f.fail(m, "SUPERGPU_OPERATION_REQUIRED", errors.New("operation is required"))
	}

	gpu, err := f.engine.Execute(ctx, "supergpu.execute@1.0.0", m.Payload, map[string]string{
		"device":         device,
		"operation":      operation,
		"correlation_id": m.CorrelationID,
	})
	if err != nil {
		return f.fail(m, "SUPERGPU_STAGE_FAILED", err)
	}

	// Mesh remains the system transport boundary. The fusion itself is local to
	// N07 ownership, while peers can invoke it through soul-mesh/1.
	meshState := map[string]any{
		"transport":         "soul-mesh/1",
		"canonical_owner":   "N07",
		"peer_client_ready": f.peers != nil,
	}
	raw, _ := json.Marshal(map[string]any{
		"orbital":    estimate.Metadata,
		"prefrontal": admission.Metadata,
		"supergpu":   gpu.Metadata,
		"mesh":       meshState,
		"elapsed_ms": time.Since(start).Milliseconds(),
	})

	return protocol.Result{
		TraceID: m.TraceID, CorrelationID: m.CorrelationID,
		Source: "N07.octacore", Target: m.Source, Status: "ok",
		Metadata: map[string]string{
			"fusion_json":        string(raw),
			"orchestrator_stage": "completed",
			"orbital_stage":      "completed",
			"prefrontal_stage":   "completed",
			"supergpu_stage":     "completed",
			"hardware_claim":     "none",
		},
	}, nil
}

func (f *Fusion) fail(m protocol.Message, code string, err error) (protocol.Result, error) {
	return protocol.Result{
		TraceID: m.TraceID, CorrelationID: m.CorrelationID,
		Source: "N07.octacore", Target: m.Source, Status: "error", Error: fmt.Sprintf("%s:%s", code, err.Error()),
	}, err
}
