package cognitive

import (
    "context"
    "errors"
    "strings"
    "time"

    "github.com/divibisoul/Orquestrador-/mesh"
    "github.com/divibisoul/Orquestrador-/octacore"
)

type OctaCoreExecutor struct {
    processor *octacore.Processor
    peers *mesh.PeerClient
}

func NewOctaCoreExecutor(processor *octacore.Processor, peers *mesh.PeerClient) (*OctaCoreExecutor, error) {
    if processor == nil { return nil, errors.New("Octacore processor is required") }
    if peers == nil { return nil, errors.New("Mesh peer client is required") }
    return &OctaCoreExecutor{processor: processor, peers: peers}, nil
}

func (e *OctaCoreExecutor) Execute(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
    capability = strings.TrimSpace(capability)
    if capability == "" || strings.TrimSpace(correlation) == "" { return nil, "", errors.New("capability and correlation are required") }
    owner, err := NewPlanner(e.peers).DiscoverTool(ctx, capability)
    if err != nil { return nil, "", err }
    step := Step{ID: octacore.NewJobID(), GoalID: "ad-hoc", Capability: capability, Target: owner.Owner, BackendPrefs: []string{string(octacore.BackendRemoteMesh)}, Kind: classifyKind(capability), Payload: cloneMap(payload), CorrelationID: correlation}
    obs := e.ExecuteStep(ctx, step)
    if !obs.OK { return obs.Output, obs.Peer, errors.New(obs.Error) }
    return obs.Output, obs.Peer, nil
}

func (e *OctaCoreExecutor) ExecuteStep(ctx context.Context, step Step) Observation {
    job := octacore.Job{JobID: step.ID, CorrelationID: step.CorrelationID, Kind: stepKind(step.Kind), Source: octacore.G7, Target: step.Target, BackendPrefs: backends(step.BackendPrefs), ParallelGroup: step.ParallelGroup, Barrier: step.Barrier, Payload: map[string]any{"capability": step.Capability, "payload": cloneMap(step.Payload)}, Priority: 50, TTLMS: 30000}
    return observationFromResult(step, e.processor.Submit(ctx, job))
}

func (e *OctaCoreExecutor) ExecuteSteps(ctx context.Context, steps []Step) []Observation {
    jobs := make([]octacore.Job, 0, len(steps))
    for _, step := range steps { jobs = append(jobs, octacore.Job{JobID: step.ID, CorrelationID: step.CorrelationID, Kind: stepKind(step.Kind), Source: octacore.G7, Target: step.Target, BackendPrefs: backends(step.BackendPrefs), ParallelGroup: step.ParallelGroup, Barrier: step.Barrier, Payload: map[string]any{"capability": step.Capability, "payload": cloneMap(step.Payload)}, Priority: 50, TTLMS: 30000}) }
    results := e.processor.Batch(ctx, jobs)
    observations := make([]Observation, len(steps))
    for i, result := range results { observations[i] = observationFromResult(steps[i], result) }
    return observations
}

func observationFromResult(step Step, result octacore.Result) Observation {
    peer := step.Target
    obs := Observation{GoalID: step.GoalID, StepID: step.ID, Capability: step.Capability, OK: result.OK, Peer: peer, BackendUsed: result.BackendUsed, LatencyMS: result.Metrics.LatencyMS, QueueWaitMS: result.Metrics.QueueWaitMS, CorrelationID: result.CorrelationID, Output: result.Output, At: nowUTC()}
    if result.Error != nil { obs.Error = result.Error.Code + ":" + result.Error.Message }
    return obs
}

func stepKind(value string) octacore.JobKind {
    switch value { case "research": return octacore.KindResearch; case "perceive": return octacore.KindPerceive; case "tool": return octacore.KindTool; case "session_step": return octacore.KindSessionStep; default: return octacore.KindCustom }
}

func backends(values []string) []octacore.Backend {
    if len(values) == 0 { return []octacore.Backend{octacore.BackendRemoteMesh} }
    out := make([]octacore.Backend, 0, len(values))
    for _, value := range values { out = append(out, octacore.Backend(strings.TrimSpace(value))) }
    return out
}