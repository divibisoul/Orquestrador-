package trinity

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrDisabled = errors.New("trinity feature flags are disabled")
	ErrRejected = errors.New("trinity prefrontal rejected workload")
	ErrUnknownTask = errors.New("unknown task type")
)

type Task struct { ID string; Kind string; Payload string; MatrixSize int; BatchSize int; MemoryNeeded float64; Precision string; Metadata map[string]string }

func (o *TrinityOrchestrator) ExecuteTask(ctx context.Context, task interface{}) (Result,error) {
	if ctx == nil { return Result{}, errors.New("nil context") }
	if o == nil || o.PFC == nil || o.Fabric == nil || o.Compute == nil { return Result{}, errors.New("trinity dependencies incomplete") }
	if !o.Config.PFCEnabled || !o.Config.FabricEnabled || !o.Config.ComputeEnabled { return Result{}, ErrDisabled }
	w,err:=adaptTaskToWorkload(task); if err!=nil{return Result{},err}
	decision,err:=o.PFC.Evaluate(ctx,w,o.Compute); if err!=nil{return Result{},fmt.Errorf("pfc evaluate: %w",err)}
	threshold:=o.Config.RiskThreshold; if threshold<=0{threshold=.75}
	if !decision.Approved || decision.ConflictScore>=threshold { return o.fallback(ctx,w,decision) }
	route,err:=o.Fabric.Route(ctx,decision.Strategy,w)
	if err!=nil { route=defaultRoute(); if o.Config.FallbackMode=="legacy" { return Result{},fmt.Errorf("route: %w",err) } }
	result,err:=o.Compute.Execute(ctx,w,route)
	if err!=nil {
		_ = o.Fabric.Feedback(ctx,Feedback{WorkloadID:w.ID,Route:route,Success:false,Error:err.Error()})
		return Result{},fmt.Errorf("compute execute: %w",err)
	}
	if trace:=traceID(ctx); trace!="" { if result.Metadata==nil {result.Metadata=map[string]string{}}; result.Metadata["trace_id"]=trace }
	if err:=o.Fabric.Feedback(ctx,Feedback{WorkloadID:w.ID,Route:route,Success:true,LatencyMS:result.LatencyMS,Quality:1});err!=nil{return Result{},fmt.Errorf("fabric feedback: %w",err)}
	if err:=o.PFC.GateWorkingMemory(ctx,w,result);err!=nil{return Result{},fmt.Errorf("working memory gate: %w",err)}
	return result,nil
}

func adaptTaskToWorkload(task interface{}) (Workload,error) {
	switch t:=task.(type) {
	case Workload: return t,validateWorkload(t)
	case *Workload: if t==nil{return Workload{},ErrUnknownTask}; return *t,validateWorkload(*t)
	case Task: return Workload{ID:t.ID,Kind:t.Kind,Payload:t.Payload,MatrixSize:t.MatrixSize,BatchSize:t.BatchSize,MemoryNeeded:t.MemoryNeeded,Precision:t.Precision,Metadata:t.Metadata},validateWorkloadFields(t.ID,t.Kind)
	case *Task: if t==nil{return Workload{},ErrUnknownTask}; return adaptTaskToWorkload(*t)
	case string: s:=strings.TrimSpace(t); if s==""{return Workload{},errors.New("empty task")}; return Workload{ID:s,Kind:"text",Payload:s,Precision:"fp32"},nil
	default: return Workload{},ErrUnknownTask
	}
}

func validateWorkload(w Workload) error { return validateWorkloadFields(w.ID,w.Kind) }
func validateWorkloadFields(id,kind string) error { if strings.TrimSpace(id)==""||strings.TrimSpace(kind)==""{return errors.New("workload id and kind required")};return nil }
func defaultRoute() Route { return Route{Target:"local",Model:"blackwell",Provider:"transcendental",Score:0.1,Fallback:"local",Capabilities:[]string{"inference","fp32"}} }
func traceID(ctx context.Context) string { if v,ok:=ctx.Value(traceKey{}).(string);ok{return v};return "" }
type traceKey struct{}
func WithTraceID(ctx context.Context,id string) context.Context { return context.WithValue(ctx,traceKey{},strings.TrimSpace(id)) }

func (o *TrinityOrchestrator) fallback(ctx context.Context,w Workload,d Decision)(Result,error){
	switch strings.ToLower(o.Config.FallbackMode){
	case "skip": return Result{Success:false,Error:ErrRejected.Error(),Metadata:map[string]string{"fallback":"skip"}},nil
	case "retry":
		route:=defaultRoute(); result,err:=o.Compute.Execute(ctx,w,route); if err!=nil{return Result{},fmt.Errorf("fallback retry: %w",err)}; _=o.Fabric.Feedback(ctx,Feedback{WorkloadID:w.ID,Route:route,Success:result.Success,LatencyMS:result.LatencyMS}); _=o.PFC.GateWorkingMemory(ctx,w,result); return result,nil
	default: return Result{Success:false,Error:ErrRejected.Error(),Metadata:map[string]string{"fallback":"legacy","reason":d.Reason}},ErrRejected
	}
}
