package compute

import (
	"context"
	"errors"
	"sort"
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

func (f *LocalFabric) Devices(context.Context) ([]Device, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]Device(nil), f.devices...), nil
}

func (f *LocalFabric) Health(context.Context) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, d := range f.devices {
		if !d.Ready {
			return errors.New("device not ready")
		}
	}
	return nil
}

func (f *LocalFabric) Execute(ctx context.Context, j Job, d Device) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if !d.Ready {
		return Result{}, errors.New("device not ready")
	}
	start := time.Now()
	work := float64(j.Tokens+1) * 1e6
	lat := work/(d.FLOPs+1)*1000 + 1
	if j.Precision == INT8 {
		lat *= .7
	}
	if j.Precision == INT4 {
		lat *= .55
	}
	if d.Utilization > .95 {
		lat *= 2
	}
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(time.Duration(lat * 1e6)):
	}
	q := .98
	if j.Precision == INT4 {
		q = .84
	}
	return Result{JobID: j.ID, DeviceID: d.ID, LatencyMs: float64(time.Since(start).Microseconds()) / 1000, EnergyJ: d.PowerWatts * lat / 1000, Quality: q, FLOPs: work}, nil
}

func (f *LocalFabric) SelectDevice(j Job) (Device, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	var best Device
	ok := false
	score := 1e99
	for _, d := range f.devices {
		if !d.Ready {
			continue
		}
		s := d.Utilization*100 + d.TemperatureC*.1 + d.PowerWatts*.01
		if j.Precision != "" {
			found := false
			for _, p := range d.Precisions {
				if p == j.Precision {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if s < score {
			score = s
			best = d
			ok = true
		}
	}
	if !ok {
		return Device{}, errors.New("no compatible device")
	}
	return best, nil
}

func (f *LocalFabric) AddDevice(d Device) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.devices = append(f.devices, d)
}

func BatchJobs(jobs []Job, size int) [][]Job {
	if size < 1 {
		size = 1
	}
	out := make([][]Job, 0, (len(jobs)+size-1)/size)
	for len(jobs) > 0 {
		n := size
		if n > len(jobs) {
			n = len(jobs)
		}
		out = append(out, jobs[:n])
		jobs = jobs[n:]
	}
	return out
}

func SortDevices(ds []Device) { sort.Slice(ds, func(i, j int) bool { return ds[i].FLOPs > ds[j].FLOPs }) }
