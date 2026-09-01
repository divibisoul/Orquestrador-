package prefrontal

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"
)

type Signal struct{ Name string; Value float64 }
type State struct{ Values map[string]float64; Updated time.Time; Goals []Goal; Context map[string]any; Constraints map[string]float64; Confidence float64 }
type Goal struct{ ID string; Name string; Text string; Urgency, Impact, Weight float64; Deadline time.Time }
type Plan struct{ ID string; Steps []string; Score float64; Confidence float64; Risk float64 }
type Cortex struct{ mu sync.RWMutex; State State; Weights map[string]float64; LastError float64; history []float64 }

func New() *Cortex { return &Cortex{State: State{Values: map[string]float64{}, Context: map[string]any{}, Constraints: map[string]float64{}}, Weights: map[string]float64{"urgency": .5, "impact": .5}} }
func NewCortex() *Cortex { return New() }

func (c *Cortex) ReadContext() State { c.mu.RLock(); defer c.mu.RUnlock(); values:=map[string]float64{};for k,v:=range c.State.Values{values[k]=v};ctx:=map[string]any{};for k,v:=range c.State.Context{ctx[k]=v};return State{Values:values,Updated:c.State.Updated,Goals:append([]Goal(nil),c.State.Goals...),Context:ctx,Constraints:copyFloatMap(c.State.Constraints),Confidence:c.State.Confidence} }
func (c *Cortex) ReadGoals() []Goal { c.mu.RLock(); defer c.mu.RUnlock(); return append([]Goal(nil), c.State.Goals...) }
func (c *Cortex) ReadConstraints() map[string]float64 { c.mu.RLock(); defer c.mu.RUnlock(); return copyFloatMap(c.State.Constraints) }
func (c *Cortex) ReadState() State { return c.ReadContext() }

func (c *Cortex) FuseSignals(signals []Signal) State { c.mu.Lock(); defer c.mu.Unlock(); if c.State.Values==nil{c.State.Values=map[string]float64{}};for _,s:=range signals{c.State.Values[s.Name]=s.Value};c.State.Updated=time.Now();return c.State }
func (c *Cortex) FuseSignalProbabilities(signals, confidence map[string]float64) float64 { var s,w float64;for k,v:=range signals{q:=confidence[k];s+=v*q;w+=q};if w==0{return 0};return s/w }
func (c *Cortex) DetectAnomalies(xs []float64) []float64 { if len(xs)<2{return nil};var sum float64;for _,x:=range xs{sum+=x};mean:=sum/float64(len(xs));var v float64;for _,x:=range xs{d:=x-mean;v+=d*d};sd:=math.Sqrt(v/float64(len(xs)));if sd==0{return nil};out:=[]float64{};for _,x:=range xs{if math.Abs(x-mean)>3*sd{out=append(out,x)}};return out }
func (c *Cortex) DetectAnomaly(x,mean,std float64) bool { return std>0&&math.Abs(x-mean)>3*std }
func (c *Cortex) CausalReason(event string, causes map[string]float64) string { best:="";score:=-math.Inf(1);for k,v:=range causes{if v>score{best,score=k,v}};if best==""{return "unknown cause for "+event};return best }
func (c *Cortex) CausalInference(event string) []string { if event=="node_failed"{return []string{"capacity_loss","route_degradation"}};return []string{"unknown"} }
func (c *Cortex) ProbabilisticReason(likelihoods map[string]float64) map[string]float64 { var total float64;for _,v:=range likelihoods{if v>0{total+=v}};out:=map[string]float64{};if total==0{return out};for k,v:=range likelihoods{if v>0{out[k]=v/total}};return out }
func (c *Cortex) ProbabilisticInference(options []Plan) []Plan { for i:=range options{options[i].Confidence=1-options[i].Risk};return options }
func (c *Cortex) GeneratePlan(goal string, steps []string) Plan { return Plan{ID:"plan-"+goal,Steps:append([]string(nil),steps...),Score:float64(len(steps))} }
func (c *Cortex) GenerateHypotheses(event string) []string { return c.CausalInference(event) }
func (c *Cortex) SimulatePlan(p Plan) map[string]float64 { return map[string]float64{"steps":float64(len(p.Steps)),"estimated_cost":float64(len(p.Steps)),"risk":1/(1+p.Score)} }
func (c *Cortex) SimulatePlanContext(_ context.Context,p Plan) Plan { p.Risk=float64(len(p.Steps))*.02;if p.Risk>1{p.Risk=1};p.Confidence=1-p.Risk;return p }
func (c *Cortex) PrioritizeGoals(gs []Goal) []Goal { c.mu.RLock();uw,iw:=c.Weights["urgency"],c.Weights["impact"];c.mu.RUnlock();sort.SliceStable(gs,func(i,j int)bool{return gs[i].Urgency*uw+gs[i].Impact*iw>gs[j].Urgency*uw+gs[j].Impact*iw});return gs }
func (c *Cortex) Decide(plans []Plan,budget float64) Plan { best:=Plan{};for _,p:=range plans{if p.Score<=budget&&p.Score>best.Score{best=p}};return best }
func (c *Cortex) ComparePlans(ps []Plan) Plan { sort.Slice(ps,func(i,j int)bool{return ps[i].Score*ps[i].Confidence>ps[j].Score*ps[j].Confidence});if len(ps)==0{return Plan{}};return ps[0] }
func (c *Cortex) SelectPlan(ps []Plan) Plan { return c.ComparePlans(c.ProbabilisticInference(ps)) }
func (c *Cortex) MakeDecision(ps []Plan) Plan { return c.SelectPlan(ps) }
func (c *Cortex) Delegate(p Plan) []string { return append([]string(nil),p.Steps...) }
func (c *Cortex) AllocateBudget(p Plan,budget float64) float64 { if budget<0{return 0};return budget/(1+float64(len(p.Steps))) }
func (c *Cortex) SetDeadlines(gs []Goal,deadline time.Time){for i:=range gs{if gs[i].Deadline.IsZero(){gs[i].Deadline=deadline}}}
func (c *Cortex) DefineSuccessCriteria(p Plan) bool{return len(p.Steps)>0&&p.Confidence>=.5}
func (c *Cortex) EvaluateOutcome(predicted,actual float64) float64{e:=math.Abs(predicted-actual);c.mu.Lock();c.LastError=e;c.mu.Unlock();return e}
func (c *Cortex) EvaluateOutcomeError(predicted,actual float64) float64{return c.EvaluateOutcome(predicted,actual)}
func (c *Cortex) LearnFromFeedback(errorValue float64){c.mu.Lock();defer c.mu.Unlock();if c.Weights==nil{c.Weights=map[string]float64{"urgency":.5,"impact":.5}};rate:=.01;delta:=math.Max(-.1,math.Min(.1,errorValue))*rate;c.Weights["urgency"]+=delta;c.Weights["impact"]-=delta}
func (c *Cortex) LearnFromMemory(item Goal){c.mu.Lock();defer c.mu.Unlock();c.State.Goals=append(c.State.Goals,item)}
func (c *Cortex) OptimizePolicy(reward float64){c.UpdatePolicies(reward)}
func (c *Cortex) UpdatePolicies(reward float64){c.mu.Lock();defer c.mu.Unlock();c.history=append(c.history,reward);if len(c.history)>10000{c.history=append([]float64(nil),c.history[len(c.history)-10000:]...)}}
func (c *Cortex) TrackExecution(reward float64){c.UpdatePolicies(reward)}
func (c *Cortex) LogPerformance(v float64){c.UpdatePolicies(v)}
func (c *Cortex) AdjustPolicy(reward float64){c.UpdatePolicies(reward)}
func (c *Cortex) OptimizePolicyAverage(rewards []float64) float64{if len(rewards)==0{return 0};var s float64;for _,r:=range rewards{s+=r};return s/float64(len(rewards))}
func (c *Cortex) UpdateBeliefs(errorValue float64){c.mu.Lock();defer c.mu.Unlock();if c.State.Confidence<0||c.State.Confidence>1||math.IsNaN(c.State.Confidence){c.State.Confidence=1};errorValue=math.Max(0,math.Min(1,errorValue));c.State.Confidence*=1-errorValue;if c.State.Confidence<0{c.State.Confidence=0}}
func (c *Cortex) ExplainDecision(p Plan) string{return "selected "+p.ID+" with score "+format(p.Score)+" based on feasibility, cost and policy weights"}
func (c *Cortex) ExplainReasoning(p Plan) string{return "plan="+p.ID+" confidence="+format(p.Confidence)+" risk="+format(p.Risk)}
func (c *Cortex) GetDecisionTrace() State{return c.ReadState()}
func (c *Cortex) ComparePredictionVsActual(pred,actual float64)float64{if pred==0{return 0};return(actual-pred)/pred}
func (c *Cortex) DetectDeviation(pred,actual,tolerance float64)bool{return math.Abs(c.ComparePredictionVsActual(pred,actual))>tolerance}
func (c *Cortex) CalculateError(pred,actual float64)float64{return math.Abs(actual-pred)}
func (c *Cortex) ComputeReward(quality,latency,energy float64)float64{return .5*quality-.3*latency-.2*energy}
func (c *Cortex) AssessProgress(done,total int)float64{if total<=0{return 1};return float64(done)/float64(total)}
func (c *Cortex) TriggerAlert(condition bool)bool{return condition}
func (c *Cortex) ReviseConstraints(k string,v float64){c.mu.Lock();defer c.mu.Unlock();if c.State.Constraints==nil{c.State.Constraints=map[string]float64{}};c.State.Constraints[k]=v}
func (c *Cortex) FineTuneObjectives(gs []Goal)[]Goal{return c.PrioritizeGoals(gs)}
func (c *Cortex) LearnFromFailure(errorValue float64){c.UpdateBeliefs(errorValue)}
func (c *Cortex) ClassifyEvent(name string)string{switch name{case"failure","timeout","thermal":return "risk";case"completed","healthy":return "normal";default:return "unknown"}}
func (c *Cortex) ExtractFeatures(s string)[]float64{return []float64{float64(len(s)),float64(len([]rune(s)))}}
func (c *Cortex) ValidateInput(s string)bool{return len(s)>0}
func (c *Cortex) AssessConfidence(p Plan)float64{return p.Confidence}

// UpdateThresholds derives the approval conflict threshold from the most recent
// measured error. Higher error tightens the gate; lower error permits measured
// recovery toward the neutral default. The result is persisted in Constraints.
func (c *Cortex) UpdateThresholds() {
	if c == nil { return }
	c.mu.Lock(); defer c.mu.Unlock()
	if c.State.Constraints == nil { c.State.Constraints = map[string]float64{} }
	threshold := c.State.Constraints["risk_threshold"]
	if threshold <= 0 || threshold > 1 || math.IsNaN(threshold) { threshold = .75 }
	errorValue := math.Max(0, math.Min(1, c.LastError))
	neutral := .75
	target := neutral - 0.5*errorValue
	threshold += 0.2 * (target - threshold)
	c.State.Constraints["risk_threshold"] = math.Max(.25, math.Min(.9, threshold))
	c.State.Updated = time.Now()
}

// ReconfigureNeuralFabric records a concrete, idempotent control request for
// the Neural Fabric integration layer. It does not mutate Fabric internals.
func (c *Cortex) ReconfigureNeuralFabric() {
	if c == nil { return }
	c.mu.Lock(); defer c.mu.Unlock()
	if c.State.Context == nil { c.State.Context = map[string]any{} }
	c.State.Context["neural_fabric.reconfigure_requested"] = true
	c.State.Context["neural_fabric.reconfigure_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Updated = time.Now()
}

// RequestReplan records an explicit replanning request consumed by the
// orchestration layer. Keeping this as state preserves a single canonical
// decision plane and avoids creating a second planner implementation here.
func (c *Cortex) RequestReplan() {
	if c == nil { return }
	c.mu.Lock(); defer c.mu.Unlock()
	if c.State.Context == nil { c.State.Context = map[string]any{} }
	c.State.Context["orchestrator.replan_requested"] = true
	c.State.Context["orchestrator.replan_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	c.State.Updated = time.Now()
}

func copyFloatMap(in map[string]float64)map[string]float64{out:=map[string]float64{};for k,v:=range in{out[k]=v};return out}
func format(v float64)string{if v==float64(int(v)){return itoa(int(v))};return "measured"}
func itoa(v int)string{if v==0{return "0"};neg:=v<0;if neg{v=-v};b:=[]byte{};for v>0{b=append([]byte{byte('0'+v%10)},b...);v/=10};if neg{b=append([]byte{'-'},b...)};return string(b)}
