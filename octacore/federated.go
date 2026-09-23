package octacore

import (
    "context"
    "errors"
    "fmt"
    "strings"
)

type FederatedContextInput struct {
    CorrelationID      string
    Input              string
    ResearchPayload    map[string]any
    PerceptionPayload  map[string]any
    AllowResearchSkip bool
    Priority           int
    TTLMS              int64
}

type FederatedContextResult struct {
    CorrelationID string
    Research     map[string]any
    Perception   map[string]any
    Audit        OctaCoreResult
    Cycle        OctaCoreResult
    Barrier      string
}

func BuildFederatedPreJobs(input FederatedContextInput) []OctaCoreJob {
    ttl := input.TTLMS
    if ttl <= 0 {
        ttl = 30_000
    }
    priority := input.Priority
    if priority <= 0 {
        priority = 90
    }
    jobs := []OctaCoreJob{
        {
            JobID: NewJobID(), CorrelationID: strings.TrimSpace(input.CorrelationID), Kind: KindResearch, Source: G6, Target: "G4",
            BackendPrefs: []Backend{BackendRemoteMesh}, ParallelGroup: ptr("pre"), Barrier: ptr("pre"),
            Payload: map[string]any{"capability": "context-orchestration", "payload": input.ResearchPayload},
            Priority: priority, TTLMS: ttl,
        },
    }
    if input.PerceptionPayload != nil {
        jobs = append(jobs, OctaCoreJob{
            JobID: NewJobID(), CorrelationID: strings.TrimSpace(input.CorrelationID), Kind: KindPerceive, Source: G6, Target: "G3",
            BackendPrefs: []Backend{BackendRemoteMesh}, ParallelGroup: ptr("pre"), Barrier: ptr("pre"),
            Payload: map[string]any{"capability": perceptionCapability(input.PerceptionPayload), "payload": input.PerceptionPayload},
            Priority: priority, TTLMS: ttl,
        })
    }
    return jobs
}

func (p *Processor) ExecuteFederatedContextCycle(ctx context.Context, input FederatedContextInput) FederatedContextResult {
    correlationID := strings.TrimSpace(input.CorrelationID)
    if correlationID == "" {
        return FederatedContextResult{
            CorrelationID: correlationID,
            Audit: failedResult(OctaCoreJob{CorrelationID: correlationID}, "CORRELATION_ID_REQUIRED", errors.New("correlation id is required"), 0, 0),
            Cycle: failedResult(OctaCoreJob{CorrelationID: correlationID}, "CORRELATION_ID_REQUIRED", errors.New("correlation id is required"), 0, 0),
            Barrier: "pre",
        }
    }
    ttl := input.TTLMS
    if ttl <= 0 {
        ttl = 30_000
    }
    jobs := BuildFederatedPreJobs(input)


    pre := p.Batch(ctx, jobs)
    var research map[string]any
    var perception map[string]any
    researchOK := true
    for _, result := range pre {
        if result.JobID == "" {
            continue
        }
        if strings.HasPrefix(result.BackendUsed, string(BackendRemoteMesh)) {
            if result.OK && result.Output != nil {
                if jobsIndex := findJobByID(jobs, result.JobID); jobsIndex >= 0 {
                    if jobs[jobsIndex].Target == "G4" {
                        research = result.Output
                    } else if jobs[jobsIndex].Target == "G3" {
                        perception = result.Output
                    }
                }
            } else if jobsIndex := findJobByID(jobs, result.JobID); jobsIndex >= 0 && jobs[jobsIndex].Target == "G4" {
                researchOK = false
            }
        }
    }
    if !researchOK && !input.AllowResearchSkip {
        err := errors.New("G4_RESEARCH_FAILED")
        return FederatedContextResult{
            CorrelationID: correlationID,
            Research: research,
            Perception: perception,
            Audit: failedResult(OctaCoreJob{CorrelationID: correlationID}, "G4_RESEARCH_FAILED", err, 0, 0),
            Cycle: failedResult(OctaCoreJob{CorrelationID: correlationID}, "G4_RESEARCH_FAILED", err, 0, 0),
            Barrier: "pre",
        }
    }

    contextPayload := map[string]any{
        "research_snippets": research,
        "perception":        perception,
        "pipeline_status":   "pre_complete",
        "correlation_id":    correlationID,
    }
    audit := p.Submit(ctx, OctaCoreJob{
        JobID: NewJobID(), CorrelationID: correlationID, Kind: KindSARAudit, Source: G6, Target: "G0",
        BackendPrefs: []Backend{BackendSARAHTTP}, Barrier: ptr("pre"),
        Payload: map[string]any{"input": input.Input},
        Priority: 100, TTLMS: ttl,
    })
    if !audit.OK {
        message := "SARA audit failed"
        if audit.Error != nil {
            message = audit.Error.Message
        }
        return FederatedContextResult{CorrelationID: correlationID, Research: research, Perception: perception, Audit: audit, Cycle: failedResult(auditJobFrom(audit), "SARA_AUDIT_FAILED", errors.New(message), 0, 0), Barrier: "pre"}
    }
    cyclePayload := map[string]any{"input": input.Input, "context": contextPayload, "cycle_id": correlationID, "audit": audit.Output}
    cycle := p.Submit(ctx, OctaCoreJob{
        JobID: NewJobID(), CorrelationID: correlationID, Kind: KindSARACycle, Source: G6, Target: "G0",
        BackendPrefs: []Backend{BackendSARAHTTP},
        Payload: cyclePayload,
        Priority: 100, TTLMS: ttl,
    })
    return FederatedContextResult{
        CorrelationID: correlationID, Research: research, Perception: perception,
        Audit: audit, Cycle: cycle, Barrier: "pre",
    }
}

func perceptionCapability(payload map[string]any) string {
    if capability, ok := payload["capability"].(string); ok && strings.TrimSpace(capability) != "" {
        return strings.TrimSpace(capability)
    }
    if _, ok := payload["data"]; ok {
        return "audio.transcribe"
    }
    return "mesh.describe"
}

func findJobByID(jobs []OctaCoreJob, id string) int {
    for i, job := range jobs {
        if job.JobID == id {
            return i
        }
    }
    return -1
}

func ptr(value string) *string { return &value }

func auditJobFrom(result OctaCoreResult) OctaCoreJob {
    return OctaCoreJob{JobID: result.JobID, CorrelationID: result.CorrelationID, Kind: KindSARAudit, Source: G6, Target: "G0", BackendPrefs: []Backend{BackendSARAHTTP}, Priority: 100, TTLMS: 1}
}

func federatedError(result FederatedContextResult) error {
    if result.Cycle.Error != nil {
        return fmt.Errorf("%s:%s", result.Cycle.Error.Code, result.Cycle.Error.Message)
    }
    return nil
}
