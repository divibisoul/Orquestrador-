package grce

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/divibisoul/Orquestrador-/grf"
)

type Dependencies struct {
	ERU           ERU
	GPUDetect     FailureDetector
	CPUDetect     FailureDetector
	DetectorMerge FailureMerger
	Neocortex     Neocortex
	ETR           EthicalGate
	GPUAnalyze    Analyzer
	CPUAnalyze    Analyzer
	AnalysisMerge AnalysisMerger
	MMD           Mediator
	ARA           Regenerator
	ITR           Innovator
	RGO           Auditor
	Horta         StateStore
	Vagus         SignalEmitter
	Mesh          MeshDistributor
}

type Result struct {
	State        grf.State
	Provenance   []grf.Provenance
	Evidence     []grf.Evidence
	Capabilities []grf.Capability
	Status       grf.EpistemicState
}

func (d Dependencies) validate() error {
	required := map[string]any{
		"ERU": d.ERU, "GPUDetect": d.GPUDetect, "CPUDetect": d.CPUDetect,
		"DetectorMerge": d.DetectorMerge, "Neocortex": d.Neocortex, "ETR": d.ETR,
		"GPUAnalyze": d.GPUAnalyze, "CPUAnalyze": d.CPUAnalyze, "AnalysisMerge": d.AnalysisMerge,
		"MMD": d.MMD, "ARA": d.ARA, "ITR": d.ITR, "RGO": d.RGO,
		"Horta": d.Horta, "Vagus": d.Vagus, "Mesh": d.Mesh,
	}
	for name, value := range required {
		if value == nil {
			return fmt.Errorf("GRCE_DEPENDENCY_MISSING:%s", name)
		}
	}
	return nil
}

func GoldenRuleCycle(ctx context.Context, d Dependencies, state grf.State, c grf.Context) (Result, error) {
	if err := d.validate(); err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}
	if err := c.Validate(); err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}
	if err := grf.ValidateInvariantSet(grf.CanonicalInvariants); err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}

	snapshot, err := d.ERU.Snapshot(ctx, state, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}

	gpuFailures, cpuFailures, err := parallelDetect(ctx, d.GPUDetect, d.CPUDetect, state, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}
	failures, err := d.DetectorMerge.Merge(ctx, gpuFailures, cpuFailures, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Status: grf.EpistemicPreserved}, err
	}
	if len(failures) == 0 {
		return Result{State: state, Status: grf.EpistemicIdle}, nil
	}

	evidence := make([]grf.Evidence, 0, len(failures))
	provenance := make([]grf.Provenance, 0, len(failures))
	characterizations := make([]grf.Characterization, 0, len(failures))
	for i, failure := range failures {
		e := grf.Evidence{
			ID:          fmt.Sprintf("%s-evidence-%d", failure.ID, i),
			FailureID:   failure.ID,
			ContextHash: c.Hash(),
			Source:      "GRCE.DETECT",
			Detail:      failure.Observed,
		}
		e.Hash = hashEvidence(e)
		evidence = append(evidence, e)
		k, err := d.Neocortex.Characterize(ctx, e, c)
		if err != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
		}
		characterizations = append(characterizations, k...)
	}

	oppositions, err := d.Neocortex.Dualize(ctx, characterizations, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	oppositions, err = d.ETR.GateOppositions(ctx, oppositions, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	for _, o := range oppositions {
		if o.Property == "" || o.RequiredBy == "" {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicUnresolved},
				errors.New("GRCE_DUAL_NOT_DECLARED")
		}
	}

	gpuAnalysis, cpuAnalysis, err := parallelAnalyze(ctx, d.GPUAnalyze, d.CPUAnalyze, oppositions, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	analysisSet, err := d.AnalysisMerge.MergeAnalysis(ctx, gpuAnalysis, cpuAnalysis, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}

	transformations := make([]grf.Transformation, 0, len(analysisSet))
	for _, analysis := range analysisSet {
		integration, err := d.MMD.Mediate(ctx, analysis, state, c)
		if err != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
		}

		transformation, err := d.ARA.Regenerate(ctx, integration, c)
		if err != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
		}

		if transformation.State.Size() < state.Size() {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved},
				errors.New("GRCE_MONOTONICITY_ABORT")
		}

		transformation, err = d.ETR.Transform(ctx, transformation, c)
		if err != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
		}

		transformation, err = d.ITR.Transform(ctx, transformation, state, c)
		if err != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
		}
		transformations = append(transformations, transformation)

		p := transformation.Provenance
		if p.ParentHash == "" {
			p.ParentHash = state.Hash()
		}
		if p.InputHash == "" {
			p.InputHash = state.Hash()
		}
		if p.OutputHash == "" {
			p.OutputHash = transformation.State.Hash()
		}
		if p.SequenceIndex == 0 {
			p.SequenceIndex = uint64(len(provenance) + 1)
		}
		provenance = append(provenance, p)
	}

	candidate, err := d.Neocortex.Form(ctx, transformations, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	candidate.ParentHash = state.Hash()
	candidate.InputHash = state.Hash()
	candidate.SequenceIndex = state.SequenceIndex + 1
	candidate.Epistemic = grf.EpistemicProjected

	candidateProvenance := grf.Provenance{ParentHash: state.Hash(), InputHash: state.Hash(), OutputHash: candidate.Hash(), SequenceIndex: candidate.SequenceIndex, Chain: []string{"F", "E", "K", "O", "A", "I", "T", "F*"}}
	provenance = append(provenance, candidateProvenance)

	if err := grf.ValidateContextAndProvenance(state, c, candidate, candidateProvenance); err != nil {
		_ = d.ERU.Remember(ctx, grf.PreserveOnFailure(state), provenance, evidence, nil)
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}

	etrOK, err := d.ETR.Validate(ctx, candidate, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	itrOK, err := d.ITR.Validate(ctx, candidate, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}
	rgoOK, err := d.RGO.Audit(ctx, state, candidate, provenance, c)
	if err != nil {
		return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, err
	}

	if !(etrOK && itrOK && rgoOK) {
		rolled, rollbackErr := d.ERU.Rollback(ctx, snapshot, c)
		if rollbackErr != nil {
			return Result{State: grf.PreserveOnFailure(state), Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved}, rollbackErr
		}
		return Result{State: rolled, Provenance: provenance, Evidence: evidence, Status: grf.EpistemicPreserved},
			errors.New("GRCE_VALIDATION_ROLLBACK")
	}

	candidate.Epistemic = grf.EpistemicActive
	candidate.OutputHash = candidate.Hash()
	candidateProvenance.OutputHash = candidate.OutputHash
	provenance[len(provenance)-1] = candidateProvenance
	newCaps := make([]grf.Capability, 0, len(transformations))
	for _, t := range transformations {
		id := t.FailureID + ":capability"
		newCaps = append(newCaps, grf.Capability{
			ID:          id,
			Description: "Capability derived from an observed failure through GRF/GRCE.",
			Genealogy:   []string{t.FailureID},
			Epistemic:   grf.EpistemicActive,
		})
	}

	if err := d.Horta.Persist(ctx, candidate, provenance, evidence, newCaps); err != nil {
		return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, err
	}
	if err := d.Vagus.Emit(ctx, candidate, provenance, newCaps); err != nil {
		return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, err
	}
	if err := d.Mesh.Distribute(ctx, candidate, c); err != nil {
		return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, err
	}
	if err := d.Neocortex.Learn(ctx, []ResultView{{Status: "ACTIVE", State: candidate}}, c); err != nil {
		return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, err
	}
	if err := d.ERU.Remember(ctx, candidate, provenance, evidence, newCaps); err != nil {
		return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, err
	}

	return Result{State: candidate, Provenance: provenance, Evidence: evidence, Capabilities: newCaps, Status: grf.EpistemicActive}, nil
}

func parallelDetect(ctx context.Context, gpu, cpu FailureDetector, state grf.State, c grf.Context) ([]grf.Failure, []grf.Failure, error) {
	var wg sync.WaitGroup
	var gpuOut, cpuOut []grf.Failure
	var gpuErr, cpuErr error
	wg.Add(2)
	go func() { defer wg.Done(); gpuOut, gpuErr = gpu.Detect(ctx, state, c) }()
	go func() { defer wg.Done(); cpuOut, cpuErr = cpu.Detect(ctx, state, c) }()
	wg.Wait()
	if gpuErr != nil {
		return nil, nil, gpuErr
	}
	if cpuErr != nil {
		return nil, nil, cpuErr
	}
	sort.SliceStable(gpuOut, func(i, j int) bool { return gpuOut[i].ID < gpuOut[j].ID })
	sort.SliceStable(cpuOut, func(i, j int) bool { return cpuOut[i].ID < cpuOut[j].ID })
	return gpuOut, cpuOut, nil
}

func parallelAnalyze(ctx context.Context, gpu, cpu Analyzer, oppositions []grf.Opposition, c grf.Context) ([]grf.Analysis, []grf.Analysis, error) {
	var wg sync.WaitGroup
	var gpuOut, cpuOut []grf.Analysis
	var gpuErr, cpuErr error
	wg.Add(2)
	go func() { defer wg.Done(); gpuOut, gpuErr = gpu.Analyze(ctx, oppositions, c) }()
	go func() { defer wg.Done(); cpuOut, cpuErr = cpu.Analyze(ctx, oppositions, c) }()
	wg.Wait()
	if gpuErr != nil {
		return nil, nil, gpuErr
	}
	if cpuErr != nil {
		return nil, nil, cpuErr
	}
	sort.SliceStable(gpuOut, func(i, j int) bool { return gpuOut[i].FailureID < gpuOut[j].FailureID })
	sort.SliceStable(cpuOut, func(i, j int) bool { return cpuOut[i].FailureID < cpuOut[j].FailureID })
	return gpuOut, cpuOut, nil
}

func hashEvidence(e grf.Evidence) string {
	h := sha256.Sum256([]byte(e.ID + "|" + e.FailureID + "|" + e.ContextHash + "|" + e.Source + "|" + e.Detail))
	return hex.EncodeToString(h[:])
}
