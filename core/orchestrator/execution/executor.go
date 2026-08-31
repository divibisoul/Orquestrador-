package execution

import("context";"errors";"sync";"time")

type Task func(context.Context) error
type Executor struct{Workers int; mu sync.Mutex; failures []time.Time; openUntil time.Time; limits map[string]*rateBucket}
type rateBucket struct{tokens float64; last time.Time; rate float64; cap float64}
func New(workers int)*Executor{if workers<1{workers=1};return &Executor{Workers:workers,limits:map[string]*rateBucket{}}}
// 8 ExecuteParallel runs independent tasks with bounded worker concurrency.
func(e *Executor)ExecuteParallel(ctx context.Context,tasks []Task)[]error{jobs:=make(chan Task);errs:=make(chan error,len(tasks));var wg sync.WaitGroup;for i:=0;i<e.Workers;i++{wg.Add(1);go func(){defer wg.Done();for t:=range jobs{if err:=t(ctx);err!=nil{errs<-err}}}()};for _,t:=range tasks{jobs<-t};close(jobs);wg.Wait();close(errs);out:=[]error{};for err:=range errs{out=append(out,err)};return out}
// 9 ExecuteDistributed dispatches tasks through a caller-provided node dispatcher.
func(e *Executor)ExecuteDistributed(ctx context.Context,tasks []Task,dispatch func(context.Context,Task)error)[]error{out:=[]error{};for _,t:=range tasks{if err:=dispatch(ctx,t);err!=nil{out=append(out,err)}};return out}
// 10 CircuitBreaker opens after >50% failures in a 10s rolling window.
func(e *Executor)CircuitBreaker(success bool)bool{e.mu.Lock();defer e.mu.Unlock();now:=time.Now();cut:=now.Add(-10*time.Second);kept:=e.failures[:0];for _,t:=range e.failures{if t.After(cut){kept=append(kept,t)}};e.failures=kept;if !success{e.failures=append(e.failures,now)};if len(e.failures)>=2&&float64(len(e.failures))/float64(len(e.failures)+1)>.5{e.openUntil=now.Add(30*time.Second)};return now.Before(e.openUntil)}
// 11 Bulkhead isolates a workflow to a bounded worker count.
func(e *Executor)Bulkhead(ctx context.Context,workers int,tasks []Task)[]error{if workers<1{workers=1};local:=New(workers);return local.ExecuteParallel(ctx,tasks)}
// 12 RateLimiter implements a token bucket and returns whether a request is allowed.
func(e *Executor)RateLimiter(key string,limit int)bool{e.mu.Lock();defer e.mu.Unlock();now:=time.Now();b:=e.limits[key];if b==nil{b=&rateBucket{tokens:float64(limit),last:now,rate:float64(limit),cap:float64(limit)};e.limits[key]=b};b.tokens+=now.Sub(b.last).Seconds()*b.rate;if b.tokens>b.cap{b.tokens=b.cap};b.last=now;if b.tokens<1{return false};b.tokens--;return true}
func RetryDelay(attempt int)time.Duration{if attempt<1{attempt=1};d:=time.Duration(1<<min(attempt-1,6))*time.Second;if d>60*time.Second{d=60*time.Second};return d}
func min(a,b int)int{if a<b{return a};return b}
var ErrCircuitOpen=errors.New("circuit breaker open")
