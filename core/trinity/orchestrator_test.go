package trinity

import (
	"context"
	"testing"
)

type pfcStub struct{ called bool; gated bool }
func(p *pfcStub)Evaluate(context.Context,Workload,ComputeEngine)(Decision,error){p.called=true;return Decision{Approved:true,Strategy:Strategy{Precision:"fp32"}},nil}
func(p *pfcStub)GateWorkingMemory(context.Context,Workload,Result)error{p.gated=true;return nil}
type fabricStub struct{ routed bool; feedbacked bool }
func(f *fabricStub)Route(context.Context,Strategy,Workload)(Route,error){f.routed=true;return defaultTestRoute(),nil}
func(f *fabricStub)Feedback(context.Context,Feedback)error{f.feedbacked=true;return nil}
type computeStub struct{}
func(computeStub)Execute(context.Context,Workload,Route)(Result,error){return Result{Success:true,Metadata:map[string]string{}},nil}
func defaultTestRoute()Route{return Route{Target:"local",Model:"blackwell",Provider:"test"}}
func TestTrinityFullFlow(t *testing.T){p:=&pfcStub{};f:=&fabricStub{};o:=&TrinityOrchestrator{PFC:p,Fabric:f,Compute:computeStub{},Config:TrinityConfig{PFCEnabled:true,FabricEnabled:true,ComputeEnabled:true}};ctx:=WithTraceID(context.Background(),"trace-123");r,err:=o.ExecuteTask(ctx,Task{ID:"1",Kind:"text",Payload:"hello"});if err!=nil||!r.Success||!p.called||!p.gated||!f.routed||!f.feedbacked{t.Fatalf("flow failed r=%+v err=%v",r,err)}}
func TestFallback(t *testing.T){o:=&TrinityOrchestrator{Config:TrinityConfig{FallbackMode:"legacy"}};_,err:=o.ExecuteTask(context.Background(),Task{ID:"1",Kind:"text"});if err!=ErrDisabled{t.Fatalf("got %v",err)}}
func TestTraceID(t *testing.T){ctx:=WithTraceID(context.Background(),"abc");if traceID(ctx)!="abc"{t.Fatal("trace id lost")}}
