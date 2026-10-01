package octacore

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/divibisoul/Orquestrador-/backend"
	"github.com/divibisoul/Orquestrador-/mesh"
	"github.com/divibisoul/Orquestrador-/supergpu"
)

type Processor struct {
	scheduler *Scheduler
	executiveMu sync.RWMutex
	executive   *ExecutiveCore
}

type Config struct {
	MaxInflight          int
	TokenCapacity        int
	TokenRefillPerSecond float64
	FailureThreshold     int
	CircuitCooldown      time.Duration
}

func DefaultConfig() Config {
	return Config{
		MaxInflight:          8,
		TokenCapacity:        8,
		TokenRefillPerSecond: 8,
		FailureThreshold:     3,
		CircuitCooldown:      30 * time.Second,
	}
}

func NewProcessor(cfg Config, compute *supergpu.Runtime, peers *mesh.PeerClient, sara *backend.SARAProxy) (*Processor, error) {
	if compute == nil {
		return nil, errors.New("Octacore requires existing N07 SuperGPU runtime")
	}
	if peers == nil {
		return nil, errors.New("Octacore requires existing N07 Mesh peer client")
	}
	return &Processor{scheduler: newScheduler(cfg, compute, peers, sara)}, nil
}



func (p *Processor) SetExecutiveCore(core *ExecutiveCore) error {
	if p == nil || p.scheduler == nil {
		return errors.New("Octacore processor is unavailable")
	}
	if core == nil {
		return errors.New("executive core is required")
	}
	p.executiveMu.Lock()
	p.executive = core
	p.executiveMu.Unlock()
	return nil
}

func (p *Processor) ExecutiveCore() *ExecutiveCore {
	if p == nil || p.scheduler == nil {
		return nil
	}
	p.executiveMu.RLock()
	defer p.executiveMu.RUnlock()
	return p.executive
}

func (p *Processor) ExecutiveExecute(ctx context.Context, req ExecutiveCoreRequest) (ExecutiveCoreResult, error) {
	core := p.ExecutiveCore()
	if core == nil {
		return ExecutiveCoreResult{}, errors.New("executive core is not connected")
	}
	return core.Execute(ctx, req)
}

func (p *Processor) Submit(ctx context.Context, job Job) Result {
	return p.scheduler.execute(ctx, job)
}

func (p *Processor) Batch(ctx context.Context, jobs []Job) []Result {
	return p.scheduler.executePlan(ctx, jobs)
}

func (p *Processor) Inventory() []Slot { return p.scheduler.inventory() }

func (p *Processor) Health() map[string]any { return p.scheduler.health() }

func (p *Processor) Throttle(level int) error { return p.scheduler.setThrottle(level) }

func (p *Processor) Halt() { p.scheduler.halt() }

func (p *Processor) Resume() { p.scheduler.resume() }

func (p *Processor) SuperGPUHealth() map[string]any { return p.scheduler.superGPUHealth() }

func (p *Processor) SetVagusPublisher(fn func(context.Context, VagusEnvelope) error) {
	p.scheduler.setPublisher(fn)
}

func (p *Processor) WaitIdle(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if p.scheduler.inflightCount() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
