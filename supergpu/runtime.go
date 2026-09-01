package supergpu

import (
	"context"
	"errors"
	"math"
	"os/exec"
	"runtime"
	"sync"
)

type Device struct { ID string; Vendor string; Name string; MemoryBytes uint64; Available bool; Backend string }
type Backend interface { Execute(context.Context, Device, string, []float64) ([]float64,error) }
type Runtime struct { mu sync.RWMutex; devices []Device; backend Backend; reserved map[string]string; closed bool }

type CPUBackend struct{}
func (CPUBackend) Execute(ctx context.Context, _ Device, op string, in []float64) ([]float64,error) {
	if ctx==nil{return nil,errors.New("context is nil")}; if len(in)==0{return nil,errors.New("empty workload")}
	out:=make([]float64,len(in)); switch op { case "identity": copy(out,in); case "relu": for i,v:=range in{if v>0{out[i]=v}}; case "square": for i,v:=range in{out[i]=v*v}; case "normalize": var s float64; for _,v:=range in{s+=v*v}; norm:=math.Sqrt(s+1e-12); for i,v:=range in{out[i]=v/norm}; default:return nil,errors.New("unsupported compute operation") }
	select{case <-ctx.Done():return nil,ctx.Err();default:return out,nil}
}

func New(backend Backend) *Runtime { if backend==nil{backend=CPUBackend{}}; return &Runtime{backend:backend,reserved:make(map[string]string)} }

func (r *Runtime) Discover() []Device {
	r.mu.Lock(); defer r.mu.Unlock(); if r.closed{return nil}
	devices:=[]Device{{ID:"cpu-0",Vendor:runtime.GOOS,Name:"host-compute",Available:true,Backend:"cpu"}}
	if _,err:=exec.LookPath("nvidia-smi");err==nil{devices=append(devices,Device{ID:"nvidia-auto",Vendor:"NVIDIA",Name:"NVIDIA GPU (driver detected)",Available:true,Backend:"nvidia"})}
	if _,err:=exec.LookPath("rocminfo");err==nil{devices=append(devices,Device{ID:"amd-auto",Vendor:"AMD",Name:"AMD GPU (ROCm detected)",Available:true,Backend:"rocm"})}
	r.devices=devices; return append([]Device(nil),devices...)
}

func (r *Runtime) Select(preferred string) (Device,error) { r.mu.RLock(); defer r.mu.RUnlock(); if r.closed{return Device{},errors.New("runtime closed")}; for _,d:=range r.devices{if d.ID==preferred&&d.Available{return d,nil}}; for _,d:=range r.devices{if d.Available{return d,nil}}; return Device{},errors.New("no compute device available") }

func (r *Runtime) Reserve(deviceID, owner string) error { if deviceID==""||owner==""{return errors.New("device and owner are required")}; r.mu.Lock(); defer r.mu.Unlock(); if r.closed{return errors.New("runtime closed")}; if current,ok:=r.reserved[deviceID];ok&&current!=owner{return errors.New("device already reserved")}; r.reserved[deviceID]=owner; return nil }

func (r *Runtime) Release(deviceID, owner string) error { r.mu.Lock(); defer r.mu.Unlock(); current,ok:=r.reserved[deviceID]; if !ok{return nil}; if current!=owner{return errors.New("reservation owner mismatch")}; delete(r.reserved,deviceID); return nil }

func (r *Runtime) Execute(ctx context.Context, device Device, operation string, input []float64) ([]float64,error) { r.mu.RLock(); closed:=r.closed; backend:=r.backend; r.mu.RUnlock(); if closed{return nil,errors.New("runtime closed")}; if !device.Available{return nil,errors.New("device unavailable")}; return backend.Execute(ctx,device,operation,input) }

func (r *Runtime) Batch(ctx context.Context, device Device, operation string, inputs [][]float64) ([][]float64,error) { if len(inputs)==0{return nil,errors.New("empty batch")}; out:=make([][]float64,len(inputs)); for i,in:=range inputs{v,err:=r.Execute(ctx,device,operation,in);if err!=nil{return nil,err};out[i]=v}; return out,nil }

func (r *Runtime) MemoryStats(device Device) map[string]any { return map[string]any{"device_id":device.ID,"memory_bytes":device.MemoryBytes,"backend":device.Backend,"available":device.Available} }

func (r *Runtime) Health() map[string]any { r.mu.RLock(); defer r.mu.RUnlock(); return map[string]any{"status":map[bool]string{true:"closed",false:"ready"}[r.closed],"devices":len(r.devices),"reservations":len(r.reserved),"backend":true} }

func (r *Runtime) Shutdown() error { r.mu.Lock(); defer r.mu.Unlock(); if r.closed{return nil}; r.closed=true; r.reserved=make(map[string]string); return nil }
