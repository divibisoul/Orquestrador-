package grce

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/divibisoul/Orquestrador-/grf"
)

type SnapshotFunc func(context.Context, grf.State, grf.Context) (grf.Artifact, error)
type DetectFunc func(context.Context, grf.State, grf.Context) ([]grf.Failure, error)
type CharacterizeFunc func(context.Context, []grf.Evidence, grf.Context) ([]grf.Characterization, error)
type DualizeFunc func(context.Context, []grf.Characterization, grf.Context) ([]grf.Opposition, error)
type GateFunc func(context.Context, []grf.Opposition, grf.Context) error
type AnalyzeFunc func(context.Context, []grf.Opposition, grf.Context) ([]grf.Artifact, error)
type MediateFunc func(context.Context, []grf.Artifact, grf.State, grf.Context) ([]grf.Artifact, error)
type TransformFunc func(context.Context, grf.Artifact, grf.State, grf.Context) (grf.Artifact, error)
type FormFunc func(context.Context, []grf.Artifact, grf.State, grf.Context) (grf.State, error)
type ValidateFunc func(context.Context, grf.State, grf.Context) (bool, error)
type RollbackFunc func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, grf.Context) error
type FeedbackFunc func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, []grf.Capability, grf.Context) error
type ExtractFunc func(context.Context, grf.State, []grf.Provenance, []grf.Evidence, grf.Context) ([]grf.Capability, error)

type Hooks struct {
	Snapshot SnapshotFunc
	DetectGPU DetectFunc
	DetectCPU DetectFunc
	Characterize CharacterizeFunc
	Dualize DualizeFunc
	EthicalGate GateFunc
	AnalyzeGPU AnalyzeFunc
	AnalyzeCPU AnalyzeFunc
	Mediate MediateFunc
	Regenerate TransformFunc
	EthicalTransform TransformFunc
	Innovate TransformFunc
	Form FormFunc
	ValidateETR ValidateFunc
	ValidateITR ValidateFunc
	ValidateRGO ValidateFunc
	Rollback RollbackFunc
	Horta FeedbackFunc
	Vagus FeedbackFunc
	Mesh FeedbackFunc
	ExtractCapabilities ExtractFunc
}

type Executor struct {
	invariants grf.InvariantSet
	hooks Hooks
}

type Result struct {
	NextState grf.State `json:"next_state"`
	Provenance []grf.Provenance `json:"provenance"`
	Evidence []grf.Evidence `json:"evidence"`
	Capabilities []grf.Capability `json:"capabilities"`
	EpistemicState grf.EpistemicState `json:"epistemic_state"`
	Outcome string `json:"outcome"`
}
const Formula = "R(n+1)=F*( T( I( A( O, C ), C ), C ), C )"
const Version = "2.0.0"

func New(invariants grf.InvariantSet, hooks Hooks) (*Executor, error) {
	if err := grf.ValidateInvariantSet(invariants); err != nil { return nil, err }
	required := []struct{name string; ok bool}{
		{"Snapshot",hooks.Snapshot!=nil},{"DetectGPU",hooks.DetectGPU!=nil},{"DetectCPU",hooks.DetectCPU!=nil},
		{"Characterize",hooks.Characterize!=nil},{"Dualize",hooks.Dualize!=nil},{"EthicalGate",hooks.EthicalGate!=nil},
		{"AnalyzeGPU",hooks.AnalyzeGPU!=nil},{"AnalyzeCPU",hooks.AnalyzeCPU!=nil},{"Mediate",hooks.Mediate!=nil},
		{"Regenerate",hooks.Regenerate!=nil},{"EthicalTransform",hooks.EthicalTransform!=nil},{"Innovate",hooks.Innovate!=nil},
		{"Form",hooks.Form!=nil},{"ValidateETR",hooks.ValidateETR!=nil},{"ValidateITR",hooks.ValidateITR!=nil},
		{"ValidateRGO",hooks.ValidateRGO!=nil},{"Rollback",hooks.Rollback!=nil},{"Horta",hooks.Horta!=nil},
		{"Vagus",hooks.Vagus!=nil},{"Mesh",hooks.Mesh!=nil},{"ExtractCapabilities",hooks.ExtractCapabilities!=nil},
	}
	for _, item := range required { if !item.ok { return nil, errors.New("GRCE hook missing: "+item.name) } }
	return &Executor{invariants:invariants,hooks:hooks},nil
}

func (e *Executor) Run(ctx context.Context, state grf.State, c grf.Context) (Result,error) {
	if ctx==nil { return Result{NextState:state,EpistemicState:grf.PRESERVED,Outcome:"ABORT_PRESERVED"},errors.New("GRCE context is nil") }
	if err:=c.Validate();err!=nil{return Result{NextState:state,EpistemicState:grf.PRESERVED,Outcome:"ABORT_PRESERVED"},err}
	if err:=grf.ValidateInvariantSet(e.invariants);err!=nil{return Result{NextState:state,EpistemicState:grf.PRESERVED,Outcome:"ABORT_PRESERVED"},err}
	parentHash,err:=state.Hash();if err!=nil{return Result{NextState:state,EpistemicState:grf.PRESERVED,Outcome:"ABORT_PRESERVED"},err}
	var prov []grf.Provenance
	var evidence []grf.Evidence
	snapshot,err:=e.hooks.Snapshot(ctx,state,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"PREFLIGHT")}
	prov=append(prov,snapshot.Provenance)

	gpu,cpu,err:=parallelDetect(ctx,e.hooks.DetectGPU,e.hooks.DetectCPU,state,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"DETECT")}
	failures:=mergeFailures(gpu,cpu)
	if len(failures)==0{return Result{NextState:state,Provenance:prov,Evidence:evidence,EpistemicState:grf.IDLE,Outcome:"IDLE"},nil}

	for i,f:=range failures{
		art:=grf.Artifact{ID:"failure:"+f.ID,Stage:"DETECT",State:f.State,Payload:map[string]any{"failure":f},Provenance:grf.Provenance{ParentHash:parentHash,SequenceIndex:c.SequenceIndex+uint64(i)+1,Stage:"DETECT",CausalFailureID:f.ID}}
		ev,evErr:=grf.NewEvidence(f,state,art,c.SequenceIndex+uint64(i)+1)
		if evErr!=nil{return e.preserve(ctx,state,prov,evidence,c,evErr,"EVIDENCE")}
		evidence=append(evidence,ev)
	}
	prov=append(prov,grf.Provenance{ParentHash:parentHash,InputHash:parentHash,OutputHash:parentHash,SequenceIndex:c.SequenceIndex+10,Stage:"EVIDENCE"})

	char,err:=e.hooks.Characterize(ctx,evidence,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"CHARACTERIZE")}
	for _,item:=range char{if item.State==grf.UNRESOLVED{return e.preserve(ctx,state,prov,evidence,c,errors.New("GRCE_CHARACTERIZATION_UNRESOLVED"),"CHARACTERIZE")}}
	opps,err:=e.hooks.Dualize(ctx,char,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"DUALIZE")}
	for _,o:=range opps{if !o.PropertyDeclared||o.NecessaryProperty==""{return e.preserve(ctx,state,prov,evidence,c,errors.New("GRCE_DUAL_NOT_DECLARED"),"DUALIZE")}}
	if err:=e.hooks.EthicalGate(ctx,opps,c);err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ETR_GATE")}
	prov=append(prov,grf.Provenance{ParentHash:parentHash,InputHash:parentHash,OutputHash:parentHash,SequenceIndex:c.SequenceIndex+30,Stage:"DUALIZE_ETR"})

	ag,ac,err:=parallelAnalyze(ctx,e.hooks.AnalyzeGPU,e.hooks.AnalyzeCPU,opps,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ANALYZE")}
	integrated,err:=e.hooks.Mediate(ctx,append(ag,ac...),state,c)
	if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"INTEGRATE")}
	prov=append(prov,grf.Provenance{ParentHash:parentHash,InputHash:parentHash,OutputHash:parentHash,SequenceIndex:c.SequenceIndex+40,Stage:"ANALYZE_INTEGRATE"})

	baseSize,err:=state.Size();if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"TRANSFORM")}
	transformed:=make([]grf.Artifact,0,len(integrated))
	for _,in:=range integrated{
		t1,err:=e.hooks.Regenerate(ctx,in,state,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ARA")}
		s1,err:=t1.Size();if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ARA")}
		if s1<baseSize{return e.preserve(ctx,state,prov,evidence,c,errors.New("GRCE_MONOTONICITY_VIOLATION"),"ARA")}
		t2,err:=e.hooks.EthicalTransform(ctx,t1,state,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ETR")}
		t3,err:=e.hooks.Innovate(ctx,t2,state,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"ITR")}
		transformed=append(transformed,t3);prov=append(prov,t1.Provenance,t2.Provenance,t3.Provenance)
	}

	candidate,err:=e.hooks.Form(ctx,transformed,state,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"FORM")}
	candidate.ParentHash=parentHash
	vet,err:=e.hooks.ValidateETR(ctx,candidate,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"VALIDATE_ETR")}
	vit,err:=e.hooks.ValidateITR(ctx,candidate,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"VALIDATE_ITR")}
	vrgo,err:=e.hooks.ValidateRGO(ctx,candidate,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"VALIDATE_RGO")}
	if !vet || !vit || !vrgo {
		// Keep the promotion gate fail-closed while preserving which independent
		// validator rejected the candidate; a generic error hid real SARA/E2E causes.
		return e.preserve(ctx, state, prov, evidence, c,
			fmt.Errorf("GRCE_VALIDATION_FAILED:ETR=%t:ITR=%t:RGO=%t", vet, vit, vrgo), "VALIDATE")
	}

	caps,err:=e.hooks.ExtractCapabilities(ctx,candidate,prov,evidence,c);if err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"EXTRACT")}
	for _,cap:=range caps{
		if cap.ID=="" || cap.Provenance.CausalFailureID=="" {
			return e.preserve(ctx,state,prov,evidence,c,errors.New("GRCE_CAPABILITY_CAUSAL_TRACE_MISSING"),"EXTRACT")
		}
	}
	if err:=validateProvenanceChain(parentHash,prov,candidate);err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"PROVENANCE")}
	candidate.EpistemicState=grf.ACTIVE
	if err:=e.hooks.Horta(ctx,candidate,prov,evidence,caps,c);err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"HORTA_FEEDBACK")}
	if err:=e.hooks.Vagus(ctx,candidate,prov,evidence,caps,c);err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"VAGUS_FEEDBACK")}
	if err:=e.hooks.Mesh(ctx,candidate,prov,evidence,caps,c);err!=nil{return e.preserve(ctx,state,prov,evidence,c,err,"MESH_FEEDBACK")}
	return Result{NextState:candidate,Provenance:prov,Evidence:evidence,Capabilities:caps,EpistemicState:grf.ACTIVE,Outcome:"ACTIVE"},nil
}

func parallelDetect(ctx context.Context,gpu, cpu DetectFunc,state grf.State,c grf.Context)([]grf.Failure,[]grf.Failure,error){
	var wg sync.WaitGroup;wg.Add(2);var g,cp []grf.Failure;var ge,ce error
	go func(){defer wg.Done();g,ge=gpu(ctx,state,c)}();go func(){defer wg.Done();cp,ce=cpu(ctx,state,c)}();wg.Wait()
	if ge!=nil{return nil,nil,ge};if ce!=nil{return nil,nil,ce};return g,cp,nil
}
func parallelAnalyze(ctx context.Context,gpu,cpu AnalyzeFunc,opps []grf.Opposition,c grf.Context)([]grf.Artifact,[]grf.Artifact,error){
	var wg sync.WaitGroup;wg.Add(2);var g,cp []grf.Artifact;var ge,ce error
	go func(){defer wg.Done();g,ge=gpu(ctx,opps,c)}();go func(){defer wg.Done();cp,ce=cpu(ctx,opps,c)}();wg.Wait()
	if ge!=nil{return nil,nil,ge};if ce!=nil{return nil,nil,ce};return g,cp,nil
}
func mergeFailures(a,b []grf.Failure)[]grf.Failure{
	seen:=map[string]grf.Failure{}
	for _,f:=range append(append([]grf.Failure{},a...),b...){if f.ID==""{continue};if old,ok:=seen[f.ID];ok{if old.State==grf.BLOCKED&&f.State!=grf.BLOCKED{seen[f.ID]=f};continue};seen[f.ID]=f}
	out:=make([]grf.Failure,0,len(seen));for _,f:=range seen{out=append(out,f)};return out
}
func validateProvenanceChain(parentHash string, prov []grf.Provenance, candidate grf.State) error {
	if parentHash=="" || candidate.ParentHash=="" || candidate.ParentHash!=parentHash {
		return errors.New("GRCE_PROVENANCE_PARENT_HASH_INVALID")
	}
	if len(prov)==0 {
		return errors.New("GRCE_PROVENANCE_EMPTY")
	}
	var previous uint64
	for _,p:=range prov{
		if p.ParentHash=="" || p.InputHash=="" || p.OutputHash=="" || p.SequenceIndex==0 || p.Stage=="" {
			return errors.New("GRCE_PROVENANCE_INCOMPLETE")
		}
		if previous>0 && p.SequenceIndex<=previous {
			return errors.New("GRCE_PROVENANCE_SEQUENCE_NOT_STRICT")
		}
		previous=p.SequenceIndex
	}
	return nil
}

func (e *Executor) preserve(ctx context.Context,state grf.State,prov []grf.Provenance,evidence []grf.Evidence,c grf.Context,err error,stage string)(Result,error){
	_ = e.hooks.Rollback(ctx,state,prov,evidence,c)
	return Result{NextState:state,Provenance:prov,Evidence:evidence,EpistemicState:grf.PRESERVED,Outcome:"ROLLBACK_PRESERVED:"+stage},fmt.Errorf("%s: %w",stage,err)
}
