package orchestrator

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type Handler func(context.Context, protocol.Message) (protocol.Result,error)
type Engine struct { mu sync.RWMutex; handlers map[string]Handler; neural *neural.Network; cortex *prefrontal.Cortex; compute *supergpu.Runtime; running atomic.Bool; sequence atomic.Uint64 }

func New(n *neural.Network, c *prefrontal.Cortex, g *supergpu.Runtime) (*Engine,error) { if n==nil||c==nil||g==nil{return nil,errors.New("all nucleus services are required")}; e:=&Engine{handlers:make(map[string]Handler),neural:n,cortex:c,compute:g}; e.running.Store(true); return e,nil }

func (e *Engine) Register(operation string, handler Handler) error { if operation==""||handler==nil{return errors.New("operation and handler are required")}; e.mu.Lock(); defer e.mu.Unlock(); if _,ok:=e.handlers[operation];ok{return errors.New("handler already registered")}; e.handlers[operation]=handler; return nil }

func (e *Engine) Route(m protocol.Message) (Handler,error) { if err:=m.Validate();err!=nil{return nil,err}; e.mu.RLock(); h:=e.handlers[m.Operation]; e.mu.RUnlock(); if h==nil{return nil,errors.New("no route for operation: "+m.Operation)}; return h,nil }

func (e *Engine) Submit(ctx context.Context, m protocol.Message) (protocol.Result,error) { if !e.running.Load(){return protocol.Result{},errors.New("orchestrator stopped")}; if ctx==nil{return protocol.Result{},errors.New("context is nil")}; m.Sequence=e.sequence.Add(1); h,err:=e.Route(m); if err!=nil{return protocol.Result{TraceID:m.TraceID,Source:"N07",Target:m.Source,Status:"rejected",Error:err.Error()},err}; return h(ctx,m) }

func (e *Engine) Execute(ctx context.Context, operation string, payload []float64, metadata map[string]string) (protocol.Result,error) { m:=protocol.NewMessage("N07","N07","command",operation,payload); m.Metadata=metadata; return e.Submit(ctx,m) }

func (e *Engine) Cancel(traceID string) error { if traceID==""{return errors.New("trace id is required")}; return nil }

func (e *Engine) Status() string { if e.running.Load(){return "ready"}; return "stopped" }

func (e *Engine) Health() map[string]any { e.mu.RLock(); handlers:=len(e.handlers); e.mu.RUnlock(); return map[string]any{"nucleus":"N07","status":e.Status(),"handlers":handlers,"neural":e.neural.Health(),"prefrontal":e.cortex.Health(),"compute":e.compute.Health()} }

func (e *Engine) Stats() map[string]any { return map[string]any{"sequence":e.sequence.Load(),"status":e.Status(),"timestamp":time.Now().UTC()} }

func (e *Engine) Shutdown(ctx context.Context) error { if ctx==nil{return errors.New("context is nil")}; e.running.Store(false); select{case <-ctx.Done():return ctx.Err();default:}; return e.compute.Shutdown() }

func (e *Engine) RegisterBuiltins() error {
	if err:=e.Register("neural.forward",func(ctx context.Context,m protocol.Message)(protocol.Result,error){v,err:=e.neural.Forward(ctx,m.Payload);return protocol.Result{TraceID:m.TraceID,Source:"N07.neural",Target:m.Source,Status:status(err),Payload:v,Error:errorText(err)},err});err!=nil{return err}
	if err:=e.Register("neural.learn",func(_ context.Context,m protocol.Message)(protocol.Result,error){half:=len(m.Payload)/2;if half==0||half*2!=len(m.Payload){return protocol.Result{},errors.New("learn payload must contain input and target halves")};err:=e.neural.Learn(m.Payload[:half],m.Payload[half:]);return protocol.Result{TraceID:m.TraceID,Source:"N07.neural",Target:m.Source,Status:status(err),Error:errorText(err)},err});err!=nil{return err}
	if err:=e.Register("compute.execute",func(ctx context.Context,m protocol.Message)(protocol.Result,error){d,err:=e.compute.Select(m.Metadata["device"]);if err!=nil{return protocol.Result{},err};v,err:=e.compute.Execute(ctx,d,m.Metadata["operation"],m.Payload);return protocol.Result{TraceID:m.TraceID,Source:"N07.gpu",Target:m.Source,Status:status(err),Payload:v,Error:errorText(err)},err});err!=nil{return err}
	return nil
}

func status(err error) string {if err!=nil{return "error"};return "ok"}
func errorText(err error) string {if err!=nil{return err.Error()};return ""}
