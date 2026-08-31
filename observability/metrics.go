package observability

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct{ value uint64 }
func (c *Counter) Inc(){atomic.AddUint64(&c.value,1)}
func (c *Counter) Value()uint64{return atomic.LoadUint64(&c.value)}

type TaskMetric struct { Tasks uint64 `json:"tasks"`; Failures uint64 `json:"failures"`; TotalLatencyMS float64 `json:"total_latency_ms"` }

type Metrics struct { Tasks Counter; Failures Counter; mu sync.RWMutex; byComponent map[string]*TaskMetric }
func NewMetrics()*Metrics{return &Metrics{byComponent:make(map[string]*TaskMetric)}}
func (m *Metrics) RecordTask(failed bool){m.Record("global",0,failed)}
func (m *Metrics) Record(component string, latency time.Duration, failed bool){if m==nil{return};m.Tasks.Inc();if failed{m.Failures.Inc()};m.mu.Lock();if m.byComponent==nil{m.byComponent=make(map[string]*TaskMetric)};v:=m.byComponent[component];if v==nil{v=&TaskMetric{};m.byComponent[component]=v};v.Tasks++;if failed{v.Failures++};v.TotalLatencyMS+=float64(latency.Microseconds())/1000;m.mu.Unlock()}
func (m *Metrics) Snapshot()map[string]TaskMetric{if m==nil{return nil};m.mu.RLock();defer m.mu.RUnlock();out:=make(map[string]TaskMetric,len(m.byComponent));for k,v:=range m.byComponent{out[k]=*v};return out}

type traceKeyType struct{}
var traceKey traceKeyType
func TraceID(ctx context.Context)string{if ctx==nil{return ""};if v:=ctx.Value(traceKey);v!=nil{if s,ok:=v.(string);ok{return s}};return ""}
func WithTraceID(ctx context.Context,id string)context.Context{if ctx==nil{ctx=context.Background()};return context.WithValue(ctx,traceKey,id)}
