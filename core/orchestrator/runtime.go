package orchestrator

import("context";"errors";"sync";"time")

type Status string
const(Pending Status="pending";Running Status="running";Completed Status="completed";Failed Status="failed";Paused Status="paused")
type Step struct{ID string; Run func(context.Context)error; Compensate func(context.Context)error; Retries int}
type Workflow struct{ID string; Steps []Step; Status Status; Current int; UpdatedAt time.Time}
type Engine struct{mu sync.RWMutex; workflows map[string]*Workflow; workers chan struct{}}
func NewEngine(workers int)*Engine{if workers<1{workers=1};return &Engine{workflows:map[string]*Workflow{},workers:make(chan struct{},workers)}}
func(e *Engine)CreateWorkflow(id string,steps []Step)error{e.mu.Lock();defer e.mu.Unlock();if id==""||len(steps)==0{return errors.New("invalid workflow")};if _,ok:=e.workflows[id];ok{return errors.New("workflow exists")};e.workflows[id]=&Workflow{ID:id,Steps:steps,Status:Pending,UpdatedAt:time.Now()};return nil}
func(e *Engine)GetWorkflowStatus(id string)(Status,error){e.mu.RLock();defer e.mu.RUnlock();w,ok:=e.workflows[id];if !ok{return "",errors.New("workflow not found")};return w.Status,nil}
func(e *Engine)ExecuteStep(ctx context.Context,id string)error{e.mu.Lock();w,ok:=e.workflows[id];if !ok{e.mu.Unlock();return errors.New("workflow not found")};if w.Current>=len(w.Steps){w.Status=Completed;e.mu.Unlock();return nil};w.Status=Running;s:=w.Steps[w.Current];w.UpdatedAt=time.Now();e.mu.Unlock();select{case e.workers<-struct{}{}:case<-ctx.Done():return ctx.Err()};defer func(){<-e.workers}();var err error;for n:=0;n<=s.Retries;n++{err=s.Run(ctx);if err==nil{break};if n<s.Retries{d:=time.Second<<n;if d>60*time.Second{d=60*time.Second};select{case<-time.After(d):case<-ctx.Done():return ctx.Err()}}};e.mu.Lock();defer e.mu.Unlock();if err!=nil{w.Status=Failed;return err};w.Current++;if w.Current==len(w.Steps){w.Status=Completed};w.UpdatedAt=time.Now();return nil}
func(e *Engine)PauseWorkflow(id string)error{e.mu.Lock();defer e.mu.Unlock();w,ok:=e.workflows[id];if !ok{return errors.New("workflow not found")};if w.Status==Running{return errors.New("cannot pause active step")};w.Status=Paused;return nil}
func(e *Engine)ResumeWorkflow(id string)error{e.mu.Lock();defer e.mu.Unlock();w,ok:=e.workflows[id];if !ok{return errors.New("workflow not found")};if w.Status!=Paused{return errors.New("workflow not paused")};w.Status=Pending;return nil}
func(e *Engine)RollbackWorkflow(ctx context.Context,id string)error{e.mu.Lock();w,ok:=e.workflows[id];if !ok{e.mu.Unlock();return errors.New("workflow not found")};done:=w.Current;w.Status=Paused;w.Current=0;e.mu.Unlock();for i:=done-1;i>=0;i--{if w.Steps[i].Compensate!=nil{if err:=w.Steps[i].Compensate(ctx);err!=nil{return err}}};return nil}
func(e *Engine)RetryFailedStep(ctx context.Context,id string)error{return e.ExecuteStep(ctx,id)}
func(e *Engine)ExecuteParallel(ctx context.Context,steps []Step)error{var wg sync.WaitGroup;errs:=make(chan error,len(steps));for _,s:=range steps{s:=s;wg.Add(1);go func(){defer wg.Done();select{case e.workers<-struct{}{}:case<-ctx.Done():errs<-ctx.Err();return};defer func(){<-e.workers}();if err:=s.Run(ctx);err!=nil{errs<-err}}()};wg.Wait();close(errs);for err:=range errs{if err!=nil{return err}};return nil}
func(e *Engine)ExecuteDistributed(ctx context.Context,steps []Step,dispatch func(context.Context,Step)error)error{var wg sync.WaitGroup;errs:=make(chan error,len(steps));for _,s:=range steps{s:=s;wg.Add(1);go func(){defer wg.Done();if err:=dispatch(ctx,s);err!=nil{errs<-err}}()};wg.Wait();close(errs);for err:=range errs{if err!=nil{return err}};return nil}
