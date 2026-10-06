package grce

import (
	"context"

	"github.com/divibisoul/Orquestrador-/grf"
)

type FailureDetector interface {
	Detect(context.Context, grf.State, grf.Context) ([]grf.Failure, error)
}

type FailureMerger interface {
	Merge(context.Context, []grf.Failure, []grf.Failure, grf.Context) ([]grf.Failure, error)
}

type Neocortex interface {
	Characterize(context.Context, grf.Evidence, grf.Context) ([]grf.Characterization, error)
	Dualize(context.Context, []grf.Characterization, grf.Context) ([]grf.Opposition, error)
	Form(context.Context, []grf.Transformation, grf.Context) (grf.State, error)
	Learn(context.Context, []grceResultView, grf.Context) error
}

type EthicalGate interface {
	GateOppositions(context.Context, []grf.Opposition, grf.Context) ([]grf.Opposition, error)
	Transform(context.Context, grf.Transformation, grf.Context) (grf.Transformation, error)
	Validate(context.Context, grf.State, grf.Context) (bool, error)
}

type Analyzer interface {
	Analyze(context.Context, []grf.Opposition, grf.Context) ([]grf.Analysis, error)
}

type AnalysisMerger interface {
	MergeAnalysis(context.Context, []grf.Analysis, []grf.Analysis, grf.Context) ([]grf.Analysis, error)
}

type Mediator interface {
	Mediate(context.Context, grf.Analysis, grf.State, grf.Context) (grf.Integration, error)
}

type Regenerator interface {
	Regenerate(context.Context, grf.Integration, grf.Context) (grf.Transformation, error)
}

type Innovator interface {
	Transform(context.Context, grf.Transformation, grf.State, grf.Context) (grf.Transformation, error)
	Validate(context.Context, grf.State, grf.Context) (bool, error)
}

type Auditor interface {
	Audit(context.Context, grf.State, grf.State, []grf.Provenance, grf.Context) (bool, error)
}

type StateStore interface {
	Persist(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability) error
}

type SignalEmitter interface {
	Emit(context.Context, grf.State, []grf.Provenance, []grf.Capability) error
}

type MeshDistributor interface {
	Distribute(context.Context, grf.State, grf.Context) error
}

type ERU interface {
	Snapshot(context.Context, grf.State, grf.Context) (string, error)
	Rollback(context.Context, string, grf.Context) (grf.State, error)
	Remember(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability) error
}

type ResultView struct {
	Status string
	State  grf.State
}
