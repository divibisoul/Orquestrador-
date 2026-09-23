package hortacore

import (
    "context"
    "github.com/divibisoul/Orquestrador-/octacore"
)

type Fusion struct {
    core *HortaCore
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
        "hortacore": true,
        "vagus_control": f.control != nil && f.control.Status() != "UNCONFIGURED",
        "mesh_execution": true,
        "octacore_scheduler": true,
    }
    return description
}

func (f *Fusion) SyncMesh(ctx context.Context, correlationID string) map[string]any {
    peers := []string{"N01", "N02", "N03", "N04", "N05", "N06"}
    results := map[string]any{}
    for _, nucleus := range peers {
        description, err := f.core.processor.DiscoverPeer(ctx, nucleus)
        if err != nil {
            results[nucleus] = map[string]any{"status": "UNAVAILABLE", "error": err.Error()}
            continue
        }
        results[nucleus] = map[string]any{"status": "DISCOVERED", "description": description}
        if f.control != nil {
            _ = f.control.Publish(ctx, octacore.VagusEnvelope{
                VagusVersion: octacore.VagusVersion,
                MessageID: octacore.NewJobID(),
                CorrelationID: correlationID,
                Source: "G7",
                Target: nucleus,
                Priority: 80,
                TTL: 10_000,
                Type: "capability.hortacore",
                Payload: map[string]any{"nucleus": nucleus, "description": description},
            })
        }
    }
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
            VagusVersion: octacore.VagusVersion,
            MessageID: octacore.NewJobID(),
            CorrelationID: correlationID,
            Source: "G7",
            Target: "G0",
            Priority: 100,
            TTL: 30_000,
            Type: eventType,
            Payload: map[string]any{"processor": processorID, "result": result},
        })
    }
    return result
}
