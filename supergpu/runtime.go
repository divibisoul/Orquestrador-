package supergpu

import (
	"context"
	"errors"
	"math"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type Device struct {
	ID           string
	Vendor       string
	Name         string
	MemoryBytes  uint64
	Available    bool
	Backend      string
	Capabilities []string
}
type Reservation struct {
	DeviceID  string
	Owner     string
	ExpiresAt time.Time
	Priority  int
	Metadata  map[string]string
}
type Backend interface { Execute(context.Context, Device, string, []float64) ([]float64, error) }
type CapabilityBackend interface { Backend; Supports(Device) bool; Capabilities(Device) []string }
type Runtime struct {
	mu sync.RWMutex
	runMu sync.Mutex
	devices []Device
	backend Backend
	reserved map[string]Reservation
	closed bool
	discoverAt time.Time
	discoveryTTL time.Duration
	running sync.WaitGroup
}
type CPUBackend struct{}
func (CPUBackend) Supports(d Device) bool { return d.Backend == "cpu" }
func (CPUBackend) Capabilities(Device) []string { return []string{"fp32", "fp64", "scalar", "vector"} }
func (CPUBackend) Execute(ctx context.Context, _ Device, op string, in []float64) ([]float64, error) {
	if ctx == nil { return nil, errors.New("context is nil") }
	if len(in) == 0 { return nil, errors.New("empty workload") }
	out := make([]float64, len(in))
	for _, v := range in { if math.IsNaN(v) || math.IsInf(v, 0) { return nil, errors.New("workload contains non-finite value") } }
	switch op {
	case "identity": copy(out, in)
	case "relu": for i, v := range in { if v > 0 { out[i] = v } }
	case "square": for i, v := range in { out[i] = v*v }
	case "normalize": var s float64; for _, v := range in { s += v*v }; norm := math.Sqrt(s+1e-12); for i, v := range in { out[i] = v/norm }
	default: return nil, errors.New("unsupported compute operation")
	}
	select { case <-ctx.Done(): return nil, ctx.Err(); default: return out,nil }
}
func New(backend Backend) *Runtime { if backend==nil { backend=CPUBackend{} }; return &Runtime{backend:backend,reserved:map[string]Reservation{},discoveryTTL:10*time.Second} }
func (r *Runtime) Discover() []Device {
	r.mu.Lock(); defer r.mu.Unlock(); if r.closed{return nil}; if time.Since(r.discoverAt)<r.discoveryTTL&&len(r.devices)>0{return append([]Device(nil),r.devices...)}
	devices:=[]Device{{ID:"cpu-0",Vendor:"host",Name:runtime.GOOS+"/"+runtime.GOARCH+" CPU",Available:true,Backend:"cpu",Capabilities:[]string{"fp32","fp64","scalar","vector"}}}
	if _,err:=exec.LookPath("nvidia-smi");err==nil{devices=append(devices,Device{ID:"nvidia-auto",Vendor:"NVIDIA",Name:"NVIDIA GPU detected; compatible backend required",Available:false,Backend:"nvidia"})}
	if _,err:=exec.LookPath("rocminfo");err==nil{devices=append(devices,Device{ID:"amd-auto",Vendor:"AMD",Name:"AMD GPU detected; compatible backend required",Available:false,Backend:"rocm"})}
	if cb,ok:=r.backend.(CapabilityBackend);ok{for i,d:=range devices{if d.Backend!="cpu"{devices[i].Available=cb.Supports(d);devices[i].Capabilities=append([]string(nil),cb.Capabilities(d)...)} }}
	r.devices=devices;r.discoverAt=time.Now();return append([]Device(nil),r.devices...)
}
func (r *Runtime) Select(preferred string)(Device,error){
	r.mu.Lock();defer r.mu.Unlock();if r.closed{return Device{},errors.New("runtime closed")};r.expireReservationsLocked();if len(r.devices)==0{return Device{},errors.New("no devices discovered")}
	if preferred!=""{for _,d:=range r.devices{if d.ID==preferred{if !d.Available{return Device{},errors.New("preferred device unavailable")};if _,busy:=r.reserved[d.ID];busy{return Device{},errors.New("preferred device reserved")};return d,nil}};return Device{},errors.New("preferred device not found")}
	for _,d:=range r.devices{if d.Backend!="cpu"&&d.Available{if _,busy:=r.reserved[d.ID];!busy{return d,nil}}}
	for _,d:=range r.devices{if d.Backend=="cpu"&&d.Available{if _,busy:=r.reserved[d.ID];!busy{return d,nil}}}
	return Device{},errors.New("no compute device available")
}
func (r *Runtime) Reserve(deviceID,owner string)error{if deviceID==""||owner==""{return errors.New("device and owner are required")};r.mu.Lock();defer r.mu.Unlock();if r.closed{return errors.New("runtime closed")};r.expireReservationsLocked();var available bool;for _,d:=range r.devices{if d.ID==deviceID{available=d.Available;break}};if !available{return errors.New("device unavailable")};if current,ok:=r.reserved[deviceID];ok&&current.Owner!=owner{return errors.New("device already reserved")};r.reserved[deviceID]=Reservation{DeviceID:deviceID,Owner:owner,ExpiresAt:time.Now().Add(30*time.Second)};return nil}
func (r *Runtime) Release(deviceID,owner string)error{r.mu.Lock();defer r.mu.Unlock();current,ok:=r.reserved[deviceID];if !ok{return nil};if current.Owner!=owner{return errors.New("reservation owner mismatch")};delete(r.reserved,deviceID);return nil}
func (r *Runtime) Execute(ctx context.Context,device Device,operation string,input []float64)([]float64,error){if ctx==nil{return nil,errors.New("context is nil")};r.runMu.Lock();r.mu.RLock();closed,backend:=r.closed,r.backend;r.mu.RUnlock();if closed{r.runMu.Unlock();return nil,errors.New("runtime closed")};if !device.Available{r.runMu.Unlock();return nil,errors.New("device unavailable")};if cb,ok:=backend.(CapabilityBackend);ok&&!cb.Supports(device){r.runMu.Unlock();return nil,errors.New("backend does not support selected device")};r.running.Add(1);r.runMu.Unlock();defer r.running.Done();return backend.Execute(ctx,device,operation,input)}
func (r *Runtime) Batch(ctx context.Context,device Device,operation string,inputs [][]float64)([][]float64,error){if ctx==nil{return nil,errors.New("context is nil")};if len(inputs)==0{return nil,errors.New("empty batch")};out:=make([][]float64,len(inputs));for i,in:=range inputs{if err:=ctx.Err();err!=nil{return nil,err};v,err:=r.Execute(ctx,device,operation,in);if err!=nil{return nil,err};out[i]=v};return out,nil}
func (r *Runtime) BatchParallel(ctx context.Context,device Device,operation string,inputs [][]float64,workers int)([][]float64,error){if ctx==nil{return nil,errors.New("context is nil")};if len(inputs)==0{return nil,errors.New("empty batch")};if workers<=0{workers=1};if workers>len(inputs){workers=len(inputs)};workCtx,cancel:=context.WithCancel(ctx);defer cancel();results:=make([][]float64,len(inputs));type job struct{index int;input []float64};jobs:=make(chan job);var wg sync.WaitGroup;var firstErr error;var errMu sync.Mutex;setErr:=func(err error){errMu.Lock();defer errMu.Unlock();if firstErr==nil{firstErr=err;cancel()}};worker:=func(){defer wg.Done();for{select{case <-workCtx.Done():return;case item,ok:=<-jobs:if !ok{return};value,err:=r.Execute(workCtx,device,operation,item.input);if err!=nil{setErr(err);return};results[item.index]=value}}};wg.Add(workers);for i:=0;i<workers;i++{go worker()};for i,input:=range inputs{select{case <-workCtx.Done():close(jobs);wg.Wait();errMu.Lock();err:=firstErr;errMu.Unlock();if err!=nil{return nil,err};if cerr:=ctx.Err();cerr!=nil{return nil,cerr};return nil,errors.New("batch cancelled");case jobs<-job{index:i,input:input}}};close(jobs);wg.Wait();errMu.Lock();defer errMu.Unlock();if firstErr!=nil{return nil,firstErr};if err:=ctx.Err();err!=nil{return nil,err};return results,nil}
func (r *Runtime) MemoryStats(device Device)map[string]any{r.mu.RLock();defer r.mu.RUnlock();return map[string]any{"device_id":device.ID,"memory_bytes":device.MemoryBytes,"memory_usage_supported":device.MemoryBytes>0,"backend":device.Backend,"available":device.Available,"accelerator":device.Backend!="cpu","capabilities":append([]string(nil),device.Capabilities...)}}
func (r *Runtime) Health()map[string]any{r.mu.RLock();defer r.mu.RUnlock();status:="ready";if r.closed{status="closed"};accelerators:=0;for _,d:=range r.devices{if d.Backend!="cpu"&&d.Available{accelerators++}};if accelerators==0&&status=="ready"{status="degraded"};return map[string]any{"status":status,"devices":len(r.devices),"accelerators":accelerators,"reservations":len(r.reserved),"backend":true,"hardware_acceleration":accelerators>0,"discovery_age_ms":time.Since(r.discoverAt).Milliseconds()}}
func (r *Runtime) Shutdown()error{r.runMu.Lock();r.mu.Lock();if r.closed{r.mu.Unlock();r.runMu.Unlock();r.running.Wait();return nil};r.closed=true;for id:=range r.reserved{delete(r.reserved,id)};r.mu.Unlock();r.runMu.Unlock();r.running.Wait();return nil}
func (r *Runtime) expireReservationsLocked(){now:=time.Now();for id,res:=range r.reserved{if !res.ExpiresAt.IsZero()&&now.After(res.ExpiresAt){delete(r.reserved,id)}}}
