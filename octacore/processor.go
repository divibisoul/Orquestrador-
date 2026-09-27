package octacore

import (
	"context"

	"github.com/divibisoul/Orquestrador-/supergpu"
)

/*
*
Processor is the canonical Octacore system processor.
It owns one N07 scheduler and exposes the eight-domain execution surface.
It is software, not silicon; it reuses the existing Mesh and SuperGPU runtime.
*/
type Processor struct {
	scheduler *OctaCoreScheduler
}

func NewProcessor(cfg SchedulerConfig, control ControlPublisher) (*Processor, error) {
	runtime := supergpu.New(nil)
	runtime.Discover()
	return NewProcessorWithRuntime(cfg, control, runtime)
}

func NewProcessorWithRuntime(cfg SchedulerConfig, control ControlPublisher, runtime *supergpu.Runtime) (*Processor, error) {
	scheduler, err := NewScheduler(cfg, control, runtime)
	if err != nil {
		return nil, err
	}
	return &Processor{scheduler: scheduler}, nil
}

func (p *Processor) Submit(ctx context.Context, job OctaCoreJob) OctaCoreResult {
	return p.scheduler.Execute(ctx, job)
}

func (p *Processor) Batch(ctx context.Context, jobs []OctaCoreJob) []OctaCoreResult {
	return p.scheduler.ExecutePlan(ctx, jobs)
}

func (p *Processor) DiscoverPeer(ctx context.Context, nucleus string) (map[string]any, error) {
	return p.scheduler.DiscoverPeer(ctx, nucleus)
}

func (p *Processor) Health() SchedulerHealth { return p.scheduler.Health() }

func (p *Processor) Inventory() []OctaCoreSlot { return p.scheduler.Inventory() }

func (p *Processor) Scheduler() *OctaCoreScheduler { return p.scheduler }

func (p *Processor) ConnectSuperGPU(runtime *supergpu.Runtime) error {
	return p.scheduler.AttachSuperGPU(runtime)
}

func (p *Processor) SuperGPUConnected() bool {
	return p.scheduler.SuperGPUConnected()
}

func (p *Processor) SetThrottle(level int) error { return p.scheduler.SetThrottle(level) }

func (p *Processor) SetThrottleWithContext(ctx context.Context, level int, correlationID string) error {
	return p.scheduler.SetThrottleWithContext(ctx, level, correlationID)
}

func (p *Processor) Halt() { _ = p.HaltWithContext(context.Background(), NewJobID()) }

func (p *Processor) HaltWithContext(ctx context.Context, correlationID string) error {
	return p.scheduler.HaltWithContext(ctx, correlationID)
}

func (p *Processor) Resume() { _ = p.ResumeWithContext(context.Background(), NewJobID()) }

func (p *Processor) ResumeWithContext(ctx context.Context, correlationID string) error {
	return p.scheduler.ResumeWithContext(ctx, correlationID)
}
