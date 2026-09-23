package hortacore

import (
	"context"
	"sync"

	"github.com/divibisoul/Orquestrador-/octacore"
)

type Fusion struct {
	core    *HortaCore
	control octacore.ControlPublisher
}

func NewFusion(processor *octacore.Processor, control octacore.ControlPublisher) (*Fusion, error) {
	core, err := New(processor)
	if err != nil {
		return nil, err
	}
	return &Fusion{core: core, control: control}, nil
}

func (f *Fusion) Describe() map[string]any {
	description := f.core.Describe()
	description["fusion"] = map[string]any{
		"hortacore":              true,
		"vagus_control":          f.control != nil && f.control.Status() != "UNCONFIGURED",
		"mesh_execution":         true,
		"mesh_transport":         "canonical-soul-mesh",
		"octacore_scheduler":     true,
		"shared_supergpu_runtime": f.core.processor.SuperGPUConnected(),
		"bidirectional_control":   true,
	}
	return description
}

func (f *Fusion) Health(ctx context.Context, correlationID string) map[string]any {
	health := map[string]any{
		"status":       "READY",
		"correlation_id": correlationID,
		"describe":     f.Describe(),
		"octacore":     f.core.processor.Health(),
		"mesh":         f.SyncMesh(ctx, correlationID),
	}
	if f.control != nil {
		health["vagus_control_status"] = f.control.Status()
	}
	return health
}

func (f *Fusion) ApplySignal(signal string, level int) (map[string]any, error) {
	switch signal {
	case "throttle", "degrade":
		if level < 0 || level > 3 {
			return nil, fmt.Errorf("HortaCore signal level must be 0..3")
		}
		if err := f.core.processor.SetThrottle(level); err != nil {
			return nil, err
		}
	case "halt":
		f.core.processor.Halt()
	case "resume":
		f.core.processor.Resume()
	default:
		return nil, fmt.Errorf("unsupported HortaCore signal: %s", signal)
	}
	state := f.core.processor.Health()
	if f.control != nil {
		_ = f.control.Publish(context.Background(), octacore.VagusEnvelope{
			VagusVersion:  octacore.VagusVersion,
			MessageID:     octacore.NewJobID(),
			CorrelationID: octacore.NewJobID(),
			Source:        "G7.HortaCore",
			Target:        "VagusBus",
			Priority:      100,
			TTL:           5_000,
			Type:          "signal.hortacore",
			Payload:       map[string]any{"signal": signal, "level": level, "health": state},
		})
	}
	return map[string]any{"signal": signal, "level": level, "health": state}, nil
}

func (f *Fusion) SyncMesh(ctx context.Context, correlationID string) map[string]any {
	peers := []string{"N01", "N02", "N03", "N04", "N05", "N06"}
	results := make(map[string]any, len(peers))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, nucleus := range peers {
		nucleus := nucleus
		wg.Add(1)
		go func() {
			defer wg.Done()
			description, err := f.core.processor.DiscoverPeer(ctx, nucleus)
			row := map[string]any{"status": "UNAVAILABLE", "error": ""}
			if err == nil {
				row = map[string]any{"status": "DISCOVERED", "description": description}
			}
			if err != nil {
				row["error"] = err.Error()
			}
			mu.Lock()
			results[nucleus] = row
			mu.Unlock()
			if f.control != nil {
				payload := map[string]any{"nucleus": nucleus, "description": description}
				if err != nil {
					payload["error"] = err.Error()
				}
				_ = f.control.Publish(ctx, octacore.VagusEnvelope{
					VagusVersion:  octacore.VagusVersion,
					MessageID:     octacore.NewJobID(),
					CorrelationID: correlationID,
					Source:        "G7.HortaCore",
					Target:        nucleus,
					Priority:      80,
					TTL:           10_000,
					Type:          "capability.hortacore",
					Payload:       payload,
				})
			}
		}()
	}
	wg.Wait()
	return results
}

func (f *Fusion) Dispatch(ctx context.Context, processorID ProcessorID, input, correlationID string) octacore.OctaCoreResult {
	result := f.core.Dispatch(ctx, processorID, input, correlationID)
	if f.control != nil {
		eventType := "hortacore.result"
		if !result.OK {
			eventType = "hortacore.error"
		}
		_ = f.control.Publish(ctx, octacore.VagusEnvelope{
			VagusVersion:  octacore.VagusVersion,
			MessageID:     octacore.NewJobID(),
			CorrelationID: correlationID,
			Source:        "G7",
			Target:        "G0",
			Priority:      100,
			TTL:           30_000,
			Type:          eventType,
			Payload:       map[string]any{"processor": processorID, "result": result},
		})
	}
	return result
}
