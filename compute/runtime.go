package compute

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type LocalFabric struct {
	mu      sync.RWMutex
	devices []Device
}

func NewLocalFabric(devices []Device) *LocalFabric {
	return &LocalFabric{devices: append([]Device(nil), devices...)}
}

func (f *LocalFabric) Devices(_ context.Context) ([]Device, error) {
	if f == nil { return nil, errors.New("nil compute fabric") }
	f.mu.RLock(); defer f.mu.RUnlock()
	out := make([]Device, len(f.devices)); copy(out, f.devices)
	for i := range out { out[i].Precisions = append([]Precision(nil), out[i].Precisions...) }
	return out, nil
}

func (f *LocalFabric) Health(_ context.Context) error {
	if f == nil { return errors.New("nil compute fabric") }
	f.mu.RLock(); defer f.mu.RUnlock()
	if len(f.devices) == 0 { return errors.New("no compute devices registered") }
	for _, d := range f.devices { if !d.Ready { return errors.New("device not ready: " + d.ID) } }
	return nil
}

// Execute performs deterministic CPU work rather than sleeping to simulate latency.
// Device FLOPs/power are treated as declared telemetry for scheduling and energy estimation.
func (f *LocalFabric) Execute(ctx context.Context, j Job, d Device) (Result, error) {
	if f == nil { return Result{}, errors.New("nil compute fabric") }
	if ctx == nil { return Result{}, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return Result{}, err }
	if strings.TrimSpace(j.ID) == "" { return Result{}, errors.New("job id required") }
	if j.Tokens < 0 { return Result{}, errors.New("tokens cannot be negative") }
	if !d.Ready { return Result{}, errors.New("device not ready") }
	start := time.Now()

	iterations := j.Tokens*64 + 1024
	if iterations < 1024 { iterations = 1024 }
	if iterations > 2_000_000 { iterations = 2_000_000 }
	seed := []byte(j.ID + "|" + j.Model + "|" + string(j.Precision))
	var digest [32]byte
	for i := 0; i < iterations; i++ {
		if i%1024 == 0 {
			if err := ctx.Err(); err != nil { return Result{}, err }
		}
		var buf [40]byte
		copy(buf[:], seed)
		binary.LittleEndian.PutUint64(buf[32:], uint64(i))
		digest = sha256.Sum256(buf[:])
		seed = digest[:]
		if i%8192 == 0 { runtime.Gosched() }
	}
	elapsedMS := float64(time.Since(start).Microseconds()) / 1000
	work := float64(iterations) * 64
	energy := 0.0
	if d.PowerWatts > 0 { energy = d.PowerWatts * elapsedMS / 1000 }
	q := 0.98
	if j.Precision == INT4 { q = 0.84 }
	return Result{JobID: j.ID, DeviceID: d.ID, LatencyMs: elapsedMS, EnergyJ: energy, Quality: q, FLOPs: work}, nil
}

func (f *LocalFabric) SelectDevice(j Job) (Device, error) {
	if f == nil { return Device{}, errors.New("nil compute fabric") }
	f.mu.RLock(); defer f.mu.RUnlock()
	var best Device; ok := false; score := 1e99
	for _, d := range f.devices {
		if !d.Ready { continue }
		s := d.Utilization*100 + d.TemperatureC*.1 + d.PowerWatts*.01
		if j.Precision != "" {
			found := false
			for _, p := range d.Precisions { if p == j.Precision { found = true; break } }
			if !found { continue }
		}
		if s < score { score = s; best = d; ok = true }
	}
	if !ok { return Device{}, errors.New("no compatible device") }
	return best, nil
}

func (f *LocalFabric) AddDevice(d Device) {
	if f == nil || strings.TrimSpace(d.ID) == "" { return }
	f.mu.Lock(); defer f.mu.Unlock()
	f.devices = append(f.devices, d)
}

func BatchJobs(jobs []Job, size int) [][]Job {
	if size < 1 { size = 1 }
	out := make([][]Job, 0, (len(jobs)+size-1)/size)
	for len(jobs) > 0 { n:=size; if n>len(jobs){n=len(jobs)}; out=append(out,jobs[:n]); jobs=jobs[n:] }
	return out
}

func SortDevices(ds []Device) { sort.Slice(ds, func(i,j int)bool{return ds[i].FLOPs>ds[j].FLOPs}) }
