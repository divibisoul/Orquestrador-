package hortacore

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/divibisoul/Orquestrador-/octacore"
)

type ProcessorID string

const (
    ProcessorCodex ProcessorID = "codex"
    ProcessorBlueprint ProcessorID = "blueprint"
    ProcessorERU ProcessorID = "eru"
    ProcessorAudit ProcessorID = "audit"
    ProcessorGuide ProcessorID = "guide"
)

type Status string

const (
    PendingRepo Status = "PENDING_REPO"
    DelegateToSARA Status = "DELEGATE_SARA"
)

type Processor struct {
    ID ProcessorID
    Status Status
    Authority string
    Reason string
}

type HortaCore struct {
    processor *octacore.Processor
}

func New(processor *octacore.Processor) (*HortaCore, error) {
    if processor == nil { return nil, errors.New("Octacore processor is required") }
    return &HortaCore{processor: processor}, nil
}

func (h *HortaCore) Inventory() []Processor {
    return []Processor{
        {ID: ProcessorCodex, Status: PendingRepo, Reason: "No authoritative Codex kernel repository is registered."},
        {ID: ProcessorBlueprint, Status: PendingRepo, Reason: "No authoritative Blueprint kernel repository is registered."},
        {ID: ProcessorERU, Status: DelegateToSARA, Authority: "SARA", Reason: "ERU authority remains in SARA; no local duplicate is allowed."},
        {ID: ProcessorAudit, Status: DelegateToSARA, Authority: "SARA", Reason: "Audit authority remains in SARA."},
        {ID: ProcessorGuide, Status: PendingRepo, Reason: "No authoritative Guide kernel repository is registered."},
    }
}

func (h *HortaCore) Describe() map[string]any {
    items := h.Inventory()
    processors := make([]map[string]any, 0, len(items))
    for _, item := range items {
        processors = append(processors, map[string]any{"id": item.ID, "status": item.Status, "authority": item.Authority, "reason": item.Reason})
    }
    return map[string]any{
        "name": "HortaCore",
        "type": "federated_composition_layer",
        "preserves_native_functions": true,
        "duplicates_regeneration": false,
        "processors": processors,
        "octacore_slot_owner": "G7",
        "control_plane": "VagusBus",
        "execution_plane": "canonical-soul-mesh",
    }
}

func (h *HortaCore) Dispatch(ctx context.Context, processorID ProcessorID, input string, correlationID string) octacore.OctaCoreResult {
    switch processorID {
    case ProcessorAudit:
        job := octacore.OctaCoreJob{
            JobID: octacore.NewJobID(),
            CorrelationID: strings.TrimSpace(correlationID),
            Kind: octacore.KindSARAudit,
            Source: octacore.G7,
            Target: "G0",
            BackendPrefs: []octacore.Backend{octacore.BackendSARAHTTP},
            Payload: map[string]any{"input": input},
            Priority: 100,
            TTLMS: 30_000,
        }
        return h.processor.Submit(ctx, job)
    case ProcessorERU:
        return octacore.OctaCoreResult{
            JobID: octacore.NewJobID(), CorrelationID: correlationID, OK: false,
            Error: &octacore.OctaCoreError{Code: "HORTACORE_ERU_DELEGATION_REQUIRES_SARA_CONTRACT", Message: "ERU must remain in SARA; no local HortaCore ERU implementation exists."},
        }
    default:
        return octacore.OctaCoreResult{
            JobID: octacore.NewJobID(), CorrelationID: correlationID, OK: false,
            Error: &octacore.OctaCoreError{Code: "HORTACORE_PROCESSOR_PENDING", Message: fmt.Sprintf("processor %s is pending its authoritative repository/contract", processorID)},
        }
    }
}
