package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/divibisoul/Orquestrador-/core/orchestrator/execution"
	"github.com/divibisoul/Orquestrador-/core/orchestrator/fractal"
)

type Status string

const (
	Pending       Status = "pending"
	Running       Status = "running"
	Completed     Status = "completed"
	Failed        Status = "failed"
	Paused        Status = "paused"
	RollingBack   Status = "rolling_back"
	RollbackFailed Status = "rollback_failed"
)

type Step struct {
	ID         string
	Run        func(context.Context) error
	Compensate func(context.Context) error
	Retries    int
}

type Workflow struct {
	ID            string
	Steps         []Step
	Status        Status
	Current       int
	UpdatedAt     time.Time
	ExecutionID   string
	RollbackError string
}

type Engine struct {
	mu         sync.RWMutex
	workflows  map[string]*Workflow
	workers    chan struct{}
	execution  *execution.Executor
	fractal    *fractal.Manager
	executionSeq atomic.Uint64
}

var (
	ErrWorkflowBusy   = errors.New("workflow step already running")
	ErrWorkflowPaused = errors.New("workflow is paused")
	ErrWorkflowEmpty  = errors.New("workflow has no steps")
	ErrInvalidStep    = errors.New("workflow contains an invalid step")
)

func NewEngine(workers int) *Engine {
	if workers < 1 {
		workers = 1
	}
	return &Engine{
		workflows: map[string]*Workflow{},
		workers:   make(chan struct{}, workers),
		execution: execution.New(workers),
		fractal:   fractal.New(),
	}
}

func (e *Engine) CreateWorkflow(id string, steps []Step) error {
	if id == "" || len(steps) == 0 {
		return errors.New("invalid workflow")
	}
	for _, step := range steps {
		if step.ID == "" || step.Run == nil || step.Retries < 0 {
			return ErrInvalidStep
		}
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.workflows[id]; ok {
		return errors.New("workflow exists")
	}
	e.workflows[id] = &Workflow{
		ID:        id,
		Steps:     append([]Step(nil), steps...),
		Status:    Pending,
		UpdatedAt: time.Now(),
	}
	return nil
}

func (e *Engine) GetWorkflowStatus(id string) (Status, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	w, ok := e.workflows[id]
	if !ok {
		return "", errors.New("workflow not found")
	}
	return w.Status, nil
}

func (e *Engine) GetWorkflow(id string) (Workflow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	w, ok := e.workflows[id]
	if !ok {
		return Workflow{}, errors.New("workflow not found")
	}
	copy := *w
	copy.Steps = append([]Step(nil), w.Steps...)
	return copy, nil
}

func (e *Engine) ExecuteStep(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return errors.New("workflow not found")
	}
	if w.Status == Paused || w.Status == RollingBack || w.Status == RollbackFailed {
		e.mu.Unlock()
		return ErrWorkflowPaused
	}
	if w.ExecutionID != "" {
		e.mu.Unlock()
		return ErrWorkflowBusy
	}
	if w.Current >= len(w.Steps) {
		w.Status = Completed
		w.UpdatedAt = time.Now()
		e.mu.Unlock()
		return nil
	}

	s := w.Steps[w.Current]
	executionID := fmt.Sprintf("wf-%d", e.executionSeq.Add(1))
	w.ExecutionID = executionID
	w.Status = Running
	w.UpdatedAt = time.Now()
	e.mu.Unlock()

	select {
	case e.workers <- struct{}{}:
	case <-ctx.Done():
		e.finishStep(id, executionID, ctx.Err(), false)
		return ctx.Err()
	}
	defer func() { <-e.workers }()

	var err error
	for attempt := 0; attempt <= s.Retries; attempt++ {
		if err = ctx.Err(); err != nil {
			e.finishStep(id, executionID, err, false)
			return err
		}
		err = s.Run(ctx)
		if err == nil {
			break
		}
		if attempt < s.Retries {
			d := execution.RetryDelay(attempt + 1)
			select {
			case <-time.After(d):
			case <-ctx.Done():
				e.finishStep(id, executionID, ctx.Err(), false)
				return ctx.Err()
			}
		}
	}

	e.finishStep(id, executionID, err, err == nil)
	return err
}

func (e *Engine) finishStep(id, executionID string, err error, success bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	w, ok := e.workflows[id]
	if !ok || w.ExecutionID != executionID {
		return
	}
	w.ExecutionID = ""
	w.UpdatedAt = time.Now()
	if !success {
		w.Status = Failed
		return
	}
	w.Current++
	if w.Current >= len(w.Steps) {
		w.Status = Completed
	} else {
		w.Status = Pending
	}
}

func (e *Engine) PauseWorkflow(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	w, ok := e.workflows[id]
	if !ok {
		return errors.New("workflow not found")
	}
	if w.ExecutionID != "" || w.Status == Running {
		return errors.New("cannot pause active step")
	}
	if w.Status == Completed {
		return errors.New("cannot pause completed workflow")
	}
	w.Status = Paused
	w.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) ResumeWorkflow(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	w, ok := e.workflows[id]
	if !ok {
		return errors.New("workflow not found")
	}
	if w.Status != Paused {
		return errors.New("workflow not paused")
	}
	w.Status = Pending
	w.UpdatedAt = time.Now()
	return nil
}

func (e *Engine) RollbackWorkflow(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("nil context")
	}

	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return errors.New("workflow not found")
	}
	if w.ExecutionID != "" {
		e.mu.Unlock()
		return ErrWorkflowBusy
	}
	done := w.Current
	steps := append([]Step(nil), w.Steps...)
	w.Status = RollingBack
	w.RollbackError = ""
	w.UpdatedAt = time.Now()
	e.mu.Unlock()

	for i := done - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			e.markRollbackFailure(id, err)
			return err
		}
		if steps[i].Compensate == nil {
			continue
		}
		if err := steps[i].Compensate(ctx); err != nil {
			e.markRollbackFailure(id, err)
			return err
		}
		e.mu.Lock()
		if current, exists := e.workflows[id]; exists {
			current.Current = i
			current.UpdatedAt = time.Now()
		}
		e.mu.Unlock()
	}

	e.mu.Lock()
	if current, exists := e.workflows[id]; exists {
		current.Current = 0
		current.Status = Paused
		current.RollbackError = ""
		current.UpdatedAt = time.Now()
	}
	e.mu.Unlock()
	return nil
}

func (e *Engine) markRollbackFailure(id string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if w, ok := e.workflows[id]; ok {
		w.Status = RollbackFailed
		w.RollbackError = err.Error()
		w.UpdatedAt = time.Now()
	}
}

func (e *Engine) RetryFailedStep(ctx context.Context, id string) error {
	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return errors.New("workflow not found")
	}
	if w.Status != Failed {
		e.mu.Unlock()
		return errors.New("workflow step is not failed")
	}
	w.Status = Pending
	w.UpdatedAt = time.Now()
	e.mu.Unlock()
	return e.ExecuteStep(ctx, id)
}

func (e *Engine) ExecuteParallel(ctx context.Context, steps []Step) error {
	tasks := make([]execution.Task, 0, len(steps))
	for _, s := range steps {
		s := s
		if s.ID == "" || s.Run == nil || s.Retries < 0 {
			return ErrInvalidStep
		}
		tasks = append(tasks, func(ctx context.Context) error {
			var err error
			for attempt := 0; attempt <= s.Retries; attempt++ {
				err = s.Run(ctx)
				if err == nil {
					return nil
				}
				if attempt < s.Retries {
					select {
					case <-time.After(execution.RetryDelay(attempt + 1)):
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
			return err
		})
	}
	for _, err := range e.execution.ExecuteParallel(ctx, tasks) {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) ExecuteDistributed(ctx context.Context, steps []Step, dispatch func(context.Context, Step) error) error {
	if dispatch == nil {
		return errors.New("nil dispatcher")
	}
	tasks := make([]execution.Task, 0, len(steps))
	for _, s := range steps {
		s := s
		if s.ID == "" || s.Run == nil || s.Retries < 0 {
			return ErrInvalidStep
		}
		tasks = append(tasks, func(ctx context.Context) error {
			return dispatch(ctx, s)
		})
	}
	for _, err := range e.execution.ExecuteDistributed(ctx, tasks, func(ctx context.Context, task execution.Task) error {
		return task(ctx)
	}) {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) CircuitBreaker(success bool) bool {
	return e.execution.CircuitBreaker(success)
}

func (e *Engine) CircuitState() string {
	return e.execution.CircuitState()
}

func (e *Engine) Bulkhead(ctx context.Context, workers int, steps []Step) error {
	tasks := make([]execution.Task, 0, len(steps))
	for _, s := range steps {
		s := s
		if s.ID == "" || s.Run == nil || s.Retries < 0 {
			return ErrInvalidStep
		}
		tasks = append(tasks, func(ctx context.Context) error { return s.Run(ctx) })
	}
	for _, err := range e.execution.Bulkhead(ctx, workers, tasks) {
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) RateLimiter(key string, limit int) bool {
	return e.execution.RateLimiter(key, limit)
}

func (e *Engine) SpawnSubOrchestrator(queue []string) *fractal.SubOrchestrator {
	return e.fractal.SpawnSubOrchestrator(queue)
}

func (e *Engine) KillSubOrchestrator(id string) bool {
	return e.fractal.KillSubOrchestrator(id)
}

func (e *Engine) RebalanceTasks() {
	e.fractal.RebalanceTasks()
}
