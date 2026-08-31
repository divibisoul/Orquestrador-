package prefrontal

import (
	"context"
	"testing"
	"github.com/divibisoul/Orquestrador-/core/trinity"
)

type estimator struct{}
func (estimator) Estimate(context.Context,trinity.Workload)(trinity.CostEstimate,error){return trinity.CostEstimate{ComputeCost:.2,Confidence:1},nil}

func TestPFCWorkingMemory(t *testing.T){p:=NewPrefrontal(trinity.PrefrontalConfig{WorkingMemoryLimit:2},nil); for i:=0;i<3;i++{w:=trinity.Workload{ID:string(rune('a'+i)),Kind:"text"}; _=p.GateWorkingMemory(context.Background(),w,trinity.Result{Success:true})}; if len(p.WorkingMemory())!=2{t.Fatalf("memory size=%d",len(p.WorkingMemory()))}}
func TestConflictMonitor(t *testing.T){m:=NewConflictMonitor(.5); if s:=m.Score(trinity.Workload{ID:"x",Kind:"text"},trinity.CostEstimate{ComputeCost:100,Confidence:.2},1); s<=0||s>1{t.Fatalf("invalid score %v",s)}}
func TestMetaRL(t *testing.T){p:=NewMetaPolicy(0); s:=p.Choose(); p.Update(s,1); if s.Name==""{t.Fatal("empty strategy")}}
func TestPFCUsesEstimator(t *testing.T){p:=NewPrefrontal(trinity.PrefrontalConfig{},estimator{}); d,err:=p.Evaluate(context.Background(),trinity.Workload{ID:"x",Kind:"text"},nil); if err!=nil||d.Estimate.Confidence!=1{t.Fatalf("estimate not used: %+v %v",d,err)}}
