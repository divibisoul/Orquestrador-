package orchestrator

import(
	"context"
	"errors"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/core/orchestrator/execution"
	"github.com/divibisoul/Orquestrador-/core/orchestrator/fractal"
)

type Status string
const(Pending Status="pending";Running Status="running";Completed Status="completed";Failed Status="failed";Paused Status="paused")
type Step struct{ID string; Run func(context.Context)error; Compensate func(context.Context)error; Retries int}
type Workflow struct{ID string; Steps []Step; Status Status; Current int; UpdatedAt time.Time}
type Engine struct{mu sync.RWMutex; workflows map[string]*Workflow; workers chan struct{}; execution *execution.Executor; fractal *fractal.Manager}
func NewEngine(workers int)*Engine{if workers<1{workers=1};return &Engine{workflows:map[string]*Workflow{},workers:make(chan struct{},workers),execution:execution.New(workers),fractal:fractal.New()}}
func(e *Engine)CreateWorkflow(id string,steps []Step)error{e.mu.Lock();defer e.mu.Unlock();if id==""||len(steps)==0{return errors.New("invalid workflow")};if _,ok:=e.workflows[id];ok{return errors.New("workflow exists")};e.workflows[id]=&Workflow{ID:id,Steps:append([]Step(nil),steps...),Status:Pending,UpdatedAt:time.Now()};return nil}
func(e *Engine)GetWorkflowStatus(id string)(Status,error){e.mu.RLock();defer e.mu.RUnlock();w,ok:=e.workflows[id];if !ok{return "",errors.New("workflow not found")};return w.Status,nil}
func(e *Engine)ExecuteStep(ctx context.Context,id string)error{e.mu.Lock();w,ok:=e.workflows[id];if !ok{e.mu.Unlock();return errors.New("workflow not found")};if w.Current>=len(w.Steps){w.Status=Completed;e.mu.Unlock();return nil};w.Status=Running;s:=w.Steps[w.Current];w.UpdatedAt=time.Now();e.mu.Unlock();select{case e.workers<-struct{}{}:case<-ctx.Done():return ctx.Err()};defer func(){<-e.workers}();var err error;for n:=0;n<=s.Retries;n++{err=s.Run(ctx);if err==nil{break};if n<s.Retries{d:=execution.RetryDelay(n+1);select{case<-time.After(d):case<-ctx.Done():return ctx.Err()}}};e.mu.Lock();defer e.mu.Unlock();if err!=nil{w.Status=Failed;return err};w.Current++;if w.Current==len(w.Steps){w.Status=Completed};w.UpdatedAt=time.Now();return nil}
func(e *Engine)PauseWorkflow(id string)error{e.mu.Lock();defer e.mu.Unlock();w,ok:=e.workflows[id];if !ok{return errors.New("workflow not found")};if w.Status==Running{return errors.New("cannot pause active step")};w.Status=Paused;return nil}
func(e *Engine)ResumeWorkflow(id string)error{e.mu.Lock();defer e.mu.Unlock();w,ok:=e.workflows[id];if !ok{return errors.New("workflow not found")};if w.Status!=Paused{return errors.New("workflow not paused")};w.Status=Pending;return nil}
func(e *Engine)RollbackWorkflow(ctx context.Context,id string)error{e.mu.Lock();w,ok:=e.workflows[id];if !ok{e.mu.Unlock();return errors.New("workflow not found")};done:=w.Current;steps:=append([]Step(nil),w.Steps...);w.Status=Paused;w.Current=0;e.mu.Unlock();for i:=done-1;i>=0;i--{if steps[i].Compensate!=nil{if err:=steps[i].Compensate(ctx);err!=nil{return err}}};return nil}
func(e *Engine)RetryFailedStep(ctx context.Context,id string)error{return e.ExecuteStep(ctx,id)}
func(e *Engine)ExecuteParallel(ctx context.Context,steps []Step)error{tasks:=make([]execution.Task,0,len(steps));for _,s:=range steps{s:=s;tasks=append(tasks,func(ctx context.Context)error{return s.Run(ctx)})};for _,err:=range e.execution.ExecuteParallel(ctx,tasks){if err!=nil{return err}};return nil}
func(e *Engine)ExecuteDistributed(ctx context.Context,steps []Step,dispatch func(context.Context,Step)error)error{tasks:=make([]execution.Task,0,len(steps));for _,s:=range steps{s:=s;tasks=append(tasks,func(ctx context.Context)error{return dispatch(ctx,s)})};for _,err:=range e.execution.ExecuteDistributed(ctx,tasks,func(ctx context.Context,t execution.Task)error{return t(ctx)}){if err!=nil{return err}};return nil}
func(e *Engine)CircuitBreaker(success bool)bool{return e.execution.CircuitBreaker(success)}
func(e *Engine)Bulkhead(ctx context.Context,workers int,steps []Step)error{tasks:=make([]execution.Task,0,len(steps));for _,s:=range steps{s:=s;tasks=append(tasks,func(ctx context.Context)error{return s.Run(ctx)})};for _,err:=range e.execution.Bulkhead(ctx,workers,tasks){if err!=nil{return err}};return nil}
func(e *Engine)RateLimiter(key string,limit int)bool{return e.execution.RateLimiter(key,limit)}
func(e *Engine)SpawnSubOrchestrator(queue []string)*fractal.SubOrchestrator{return e.fractal.SpawnSubOrchestrator(queue)}
func(e *Engine)KillSubOrchestrator(id string)bool{return e.fractal.KillSubOrchestrator(id)}
func(e *Engine)RebalanceTasks(){e.fractal.RebalanceTasks()}
