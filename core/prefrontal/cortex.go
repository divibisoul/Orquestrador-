package prefrontal

import("math";"sort";"sync";"time")
type Signal struct{Name string;Value float64}
type State struct{Values map[string]float64; Updated time.Time}
type Goal struct{Name string;Urgency,Impact,Weight float64}
type Plan struct{ID string; Steps []string; Score float64}
type Cortex struct{mu sync.RWMutex; State State; Weights map[string]float64; LastError float64}
func New()*Cortex{return &Cortex{State:State{Values:map[string]float64{}},Weights:map[string]float64{"urgency":.5,"impact":.5}}}
// 16 ReadContext returns a snapshot of executive state.
func(c *Cortex)ReadContext()State{c.mu.RLock();defer c.mu.RUnlock();v:=map[string]float64{};for k,x:=range c.State.Values{v[k]=x};return State{Values:v,Updated:c.State.Updated}}
// 17 FuseSignals combines normalized signals into the state representation.
func(c *Cortex)FuseSignals(s []Signal)State{c.mu.Lock();defer c.mu.Unlock();for _,x:=range s{c.State.Values[x.Name]=x.Value};c.State.Updated=time.Now();return c.State}
// 18 DetectAnomalies uses a 3-sigma threshold over supplied observations.
func(c *Cortex)DetectAnomalies(xs []float64)[]float64{if len(xs)<2{return nil};var sum float64;for _,x:=range xs{sum+=x};mean:=sum/float64(len(xs));var v float64;for _,x:=range xs{d:=x-mean;v+=d*d};sd:=math.Sqrt(v/float64(len(xs)));if sd==0{return nil};out:=[]float64{};for _,x:=range xs{if math.Abs(x-mean)>3*sd{out=append(out,x)}};return out}
// 19 CausalReason ranks causes by observed dependency weight.
func(c *Cortex)CausalReason(event string,causes map[string]float64)string{best:="";score:=-math.Inf(1);for k,v:=range causes{if v>score{best,score=k,v}};if best==""{return "unknown cause for "+event};return best}
// 20 ProbabilisticReason normalizes plan likelihoods into posterior probabilities.
func(c *Cortex)ProbabilisticReason(likelihoods map[string]float64)map[string]float64{var total float64;for _,v:=range likelihoods{if v>0{total+=v}};out:=map[string]float64{};if total==0{return out};for k,v:=range likelihoods{if v>0{out[k]=v/total}};return out}
// 21 GeneratePlan creates a deterministic DAG-like plan from an objective.
func(c *Cortex)GeneratePlan(goal string,steps []string)Plan{return Plan{ID:"plan-"+goal,Steps:append([]string(nil),steps...),Score:float64(len(steps))}}
// 22 SimulatePlan estimates cost/risk without executing external effects.
func(c *Cortex)SimulatePlan(p Plan)map[string]float64{return map[string]float64{"steps":float64(len(p.Steps)),"estimated_cost":float64(len(p.Steps))*1.0,"risk":1/(1+p.Score)}}
// 23 PrioritizeGoals ranks goals using urgency and impact weights.
func(c *Cortex)PrioritizeGoals(gs []Goal)[]Goal{c.mu.RLock();uw,iw:=c.Weights["urgency"],c.Weights["impact"];c.mu.RUnlock();sort.SliceStable(gs,func(i,j int)bool{return gs[i].Urgency*uw+gs[i].Impact*iw>gs[j].Urgency*uw+gs[j].Impact*iw});return gs}
// 24 Decide selects the highest-scoring feasible plan.
func(c *Cortex)Decide(plans []Plan,budget float64)Plan{best:=Plan{};for _,p:=range plans{if p.Score<=budget&&p.Score>best.Score{best=p}};return best}
// 25 Delegate serializes the selected plan into executable task identifiers.
func(c *Cortex)Delegate(p Plan)[]string{return append([]string(nil),p.Steps...)}
// 26 EvaluateOutcome computes absolute prediction error.
func(c *Cortex)EvaluateOutcome(predicted,actual float64)float64{e:=math.Abs(predicted-actual);c.mu.Lock();c.LastError=e;c.mu.Unlock();return e}
// 27 LearnFromFeedback applies a bounded error correction to policy weights.
func(c *Cortex)LearnFromFeedback(errorValue float64){c.mu.Lock();defer c.mu.Unlock();rate:=.01;delta:=math.Max(-.1,math.Min(.1,errorValue))*rate;c.Weights["urgency"]+=delta;c.Weights["impact"]-=delta}
// 28 OptimizePolicy nudges weights toward a supplied reward signal.
func(c *Cortex)OptimizePolicy(reward float64){c.mu.Lock();defer c.mu.Unlock();step:=math.Max(-.05,math.Min(.05,reward*.01));c.Weights["urgency"]+=step;c.Weights["impact"]+=step}
// 29 ExplainDecision returns a concise evidence-oriented explanation.
func(c *Cortex)ExplainDecision(p Plan)string{return "selected "+p.ID+" with score "+format(p.Score)+" based on feasibility, cost and policy weights"}
func format(v float64)string{if v==float64(int(v)){return itoa(int(v))};return "measured"}
func itoa(v int)string{if v==0{return "0"};neg:=v<0;if neg{v=-v};b:=[]byte{};for v>0{b=append([]byte{byte('0'+v%10)},b...);v/=10};if neg{b=append([]byte{'-'},b...)};return string(b)}
