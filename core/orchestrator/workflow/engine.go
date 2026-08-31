package workflow

import (
 "errors"
 "fmt"
 "sync"
 "time"
)

type Status string
const ( Pending Status="Pending"; Running Status="Running"; Completed Status="Completed"; Failed Status="Failed"; Paused Status="Paused" )
type WorkflowSpec struct { Goal string; Steps []string }
type Workflow struct { ID string; Spec WorkflowSpec; Status Status; Current int; Checkpoint int; mu sync.Mutex }
type Engine struct { mu sync.RWMutex; workflows map[string]*Workflow; seq uint64 }
func NewEngine()*Engine{return &Engine{workflows:map[string]*Workflow{}}}
// 1 CreateWorkflow validates and registers a workflow DAG representation.
func(e *Engine)CreateWorkflow(s WorkflowSpec)(string,error){if len(s.Steps)==0{return "",errors.New("workflow requires steps")};e.mu.Lock();defer e.mu.Unlock();e.seq++;id:=fmt.Sprintf("wf-%d",e.seq);e.workflows[id]=&Workflow{ID:id,Spec:s,Status:Pending};return id,nil}
// 2 ExecuteStep atomically advances one workflow step.
func(e *Engine)ExecuteStep(id string)error{e.mu.RLock();w:=e.workflows[id];e.mu.RUnlock();if w==nil{return errors.New("workflow not found")};w.mu.Lock();defer w.mu.Unlock();if w.Status==Paused{return errors.New("workflow paused")};if w.Current>=len(w.Spec.Steps){w.Status=Completed;return nil};w.Status=Running;w.Current++;w.Checkpoint=w.Current;if w.Current==len(w.Spec.Steps){w.Status=Completed};return nil}
// 3 GetWorkflowStatus returns the current state.
func(e *Engine)GetWorkflowStatus(id string)Status{e.mu.RLock();w:=e.workflows[id];e.mu.RUnlock();if w==nil{return Failed};w.mu.Lock();defer w.mu.Unlock();return w.Status}
// 4 PauseWorkflow pauses at the current atomic checkpoint.
func(e *Engine)PauseWorkflow(id string)error{w:=e.get(id);if w==nil{return errors.New("workflow not found")};w.mu.Lock();defer w.mu.Unlock();if w.Status!=Running{return errors.New("workflow is not running")};w.Status=Paused;w.Checkpoint=w.Current;return nil}
// 5 ResumeWorkflow resumes from the last checkpoint.
func(e *Engine)ResumeWorkflow(id string)error{w:=e.get(id);if w==nil{return errors.New("workflow not found")};w.mu.Lock();defer w.mu.Unlock();if w.Status!=Paused{return errors.New("workflow is not paused")};w.Status=Running;return nil}
// 6 RollbackWorkflow applies local Saga-style compensation by restoring checkpoint.
func(e *Engine)RollbackWorkflow(id string,checkpoint int)error{w:=e.get(id);if w==nil{return errors.New("workflow not found")};w.mu.Lock();defer w.mu.Unlock();if checkpoint<0||checkpoint>w.Current{return errors.New("invalid checkpoint")};w.Current=checkpoint;w.Checkpoint=checkpoint;w.Status=Paused;return nil}
// 7 RetryFailedStep retries with bounded exponential backoff.
func(e *Engine)RetryFailedStep(id string,attempt int)error{if attempt<1{attempt=1};d:=time.Duration(1<<min(attempt-1,6))*time.Second;if d>60*time.Second{d=60*time.Second};time.Sleep(d);w:=e.get(id);if w==nil{return errors.New("workflow not found")};w.mu.Lock();w.Status=Running;w.mu.Unlock();return e.ExecuteStep(id)}
func(e *Engine)get(id string)*Workflow{e.mu.RLock();defer e.mu.RUnlock();return e.workflows[id]}
func min(a,b int)int{if a<b{return a};return b}
