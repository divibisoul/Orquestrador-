package orchestrator

import (
	"context"
	"encoding/json"
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
	Pending        Status = "pending"
	Running        Status = "running"
	Completed      Status = "completed"
	Failed         Status = "failed"
	Paused         Status = "paused"
	RollingBack    Status = "rolling_back"
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

type workflowPersisted struct {
	ID            string    `json:"id"`
	StepIDs       []string  `json:"step_ids"`
	Retries       []int     `json:"retries"`
	Status        Status    `json:"status"`
	Current       int       `json:"current"`
	UpdatedAt     time.Time `json:"updated_at"`
	RollbackError string    `json:"rollback_error,omitempty"`
}

type DurableStore interface {
	Put(string, []byte) (uint64, error)
	Get(string) ([]byte, uint64, bool)
}

var (
	ErrWorkflowBusy   = errors.New("workflow step already running")
	ErrWorkflowPaused = errors.New("workflow is paused")
	ErrWorkflowEmpty  = errors.New("workflow has no steps")
	ErrInvalidStep    = errors.New("workflow contains an invalid step")
)

type Engine struct {
	mu           sync.RWMutex
	workflows    map[string]*Workflow
	workers      chan struct{}
	execution    *execution.Executor
	fractal      *fractal.Manager
	executionSeq atomic.Uint64
	store        DurableStore
}

func NewEngine(workers int) *Engine {
	if workers < 1 {
		workers = 1
	}
	return &Engine{workflows: map[string]*Workflow{}, workers: make(chan struct{}, workers), execution: execution.New(workers), fractal: fractal.New()}
}

func NewEngineWithStore(workers int, store DurableStore) *Engine {
	e := NewEngine(workers)
	e.store = store
	return e
}

func (e *Engine) AttachDurableStore(store DurableStore) *Engine {
	if e == nil {
		return e
	}
	e.mu.Lock()
	e.store = store
	e.mu.Unlock()
	return e
}

func (e *Engine) persistWorkflow(w *Workflow) error {
	if e == nil || w == nil {
		return errors.New("invalid durable workflow")
	}
	e.mu.RLock()
	store := e.store
	// Snapshot every mutable workflow field while holding the engine read lock.
	// Callers may pass a live workflow pointer, so reading it after unlocking
	// would race with ExecuteStep/Pause/Resume/Rollback.
	p := workflowPersisted{
		ID:            w.ID,
		Status:        w.Status,
		Current:       w.Current,
		UpdatedAt:     w.UpdatedAt,
		RollbackError: w.RollbackError,
		StepIDs:       make([]string, len(w.Steps)),
		Retries:       make([]int, len(w.Steps)),
	}
	for i, s := range w.Steps {
		p.StepIDs[i] = s.ID
		p.Retries[i] = s.Retries
	}
	e.mu.RUnlock()
	if store == nil {
		return nil
	}
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal workflow state: %w", err)
	}
	if _, err = store.Put("workflow/"+w.ID, b); err != nil {
		return fmt.Errorf("persist workflow %s: %w", w.ID, err)
	}
	return nil
}

func (e *Engine) CreateWorkflow(id string, steps []Step) error {
	if e == nil {
		return errors.New("nil engine")
	}
	if id == "" || len(steps) == 0 {
		return ErrWorkflowEmpty
	}
	for _, step := range steps {
		if step.ID == "" || step.Run == nil || step.Retries < 0 {
			return ErrInvalidStep
		}
	}
	e.mu.Lock()
	if _, ok := e.workflows[id]; ok {
		e.mu.Unlock()
		return errors.New("workflow exists")
	}
	w := &Workflow{ID: id, Steps: append([]Step(nil), steps...), Status: Pending, UpdatedAt: time.Now()}
	e.workflows[id] = w
	e.mu.Unlock()
	if err := e.persistWorkflow(w); err != nil {
		e.mu.Lock()
		delete(e.workflows, id)
		e.mu.Unlock()
		return err
	}
	return nil
}

func (e *Engine) RestoreWorkflow(id string, steps []Step) error {
	if e == nil {
		return errors.New("nil engine")
	}
	if len(steps) == 0 {
		return ErrWorkflowEmpty
	}
	e.mu.RLock()
	store := e.store
	e.mu.RUnlock()
	if store == nil {
		return errors.New("durable store not attached")
	}
	data, _, ok := store.Get("workflow/" + id)
	if !ok {
		return errors.New("workflow state not found")
	}
	var p workflowPersisted
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("decode workflow state: %w", err)
	}
	if len(p.StepIDs) != len(steps) {
		return errors.New("workflow step definition mismatch")
	}
	for i, s := range steps {
		if s.ID == "" || s.Run == nil || s.ID != p.StepIDs[i] {
			return ErrInvalidStep
		}
	}
	w := &Workflow{ID: p.ID, Steps: append([]Step(nil), steps...), Status: p.Status, Current: p.Current, UpdatedAt: p.UpdatedAt, RollbackError: p.RollbackError}
	if w.Status == Running {
		w.Status = Pending
	}
	if w.Current < 0 || w.Current > len(w.Steps) {
		return errors.New("invalid persisted workflow cursor")
	}
	e.mu.Lock()
	e.workflows[id] = w
	e.mu.Unlock()
	return nil
}

func (e *Engine) GetWorkflowStatus(id string) (Status, error) {
	if e == nil {
		return "", errors.New("nil engine")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	w, ok := e.workflows[id]
	if !ok {
		return "", errors.New("workflow not found")
	}
	return w.Status, nil
}

func (e *Engine) GetWorkflow(id string) (Workflow, error) {
	if e == nil {
		return Workflow{}, errors.New("nil engine")
	}
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
	if e == nil {
		return errors.New("nil engine")
	}
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
		snapshot := cloneWorkflow(w)
		e.mu.Unlock()
		return e.persistWorkflow(&snapshot)
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
			timer := time.NewTimer(d)
			select {
			case <-timer.C:
			case <-ctx.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
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
	w, ok := e.workflows[id]
	if !ok || w.ExecutionID != executionID {
		e.mu.Unlock()
		return
	}
	w.ExecutionID = ""
	w.UpdatedAt = time.Now()
	if !success {
		w.Status = Failed
		w.RollbackError = ""
	} else {
		w.Current++
		if w.Current >= len(w.Steps) {
			w.Status = Completed
		} else {
			w.Status = Pending
		}
	}
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	_ = e.persistWorkflow(&snapshot)
}

func cloneWorkflow(w *Workflow) Workflow {
	copy := *w
	copy.Steps = append([]Step(nil), w.Steps...)
	return copy
}

func (e *Engine) PauseWorkflow(id string) error {
	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return errors.New("workflow not found")
	}
	if w.ExecutionID != "" || w.Status == Running {
		e.mu.Unlock()
		return errors.New("cannot pause active step")
	}
	if w.Status == Completed {
		e.mu.Unlock()
		return errors.New("cannot pause completed workflow")
	}
	w.Status = Paused
	w.UpdatedAt = time.Now()
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	return e.persistWorkflow(&snapshot)
}

func (e *Engine) ResumeWorkflow(id string) error {
	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return errors.New("workflow not found")
	}
	if w.Status != Paused {
		e.mu.Unlock()
		return errors.New("workflow not paused")
	}
	w.Status = Pending
	w.UpdatedAt = time.Now()
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	return e.persistWorkflow(&snapshot)
}

func (e *Engine) RollbackWorkflow(ctx context.Context, id string) error {
	if e == nil {
		return errors.New("nil engine")
	}
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
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	if err := e.persistWorkflow(&snapshot); err != nil {
		return err
	}
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
			snapshot = cloneWorkflow(current)
		}
		e.mu.Unlock()
		if err := e.persistWorkflow(&snapshot); err != nil {
			e.markRollbackFailure(id, err)
			return err
		}
	}
	e.mu.Lock()
	if current, exists := e.workflows[id]; exists {
		current.Current = 0
		current.Status = Paused
		current.RollbackError = ""
		current.UpdatedAt = time.Now()
		snapshot = cloneWorkflow(current)
	}
	e.mu.Unlock()
	return e.persistWorkflow(&snapshot)
}

func (e *Engine) markRollbackFailure(id string, err error) {
	e.mu.Lock()
	w, ok := e.workflows[id]
	if !ok {
		e.mu.Unlock()
		return
	}
	w.Status = RollbackFailed
	w.RollbackError = err.Error()
	w.UpdatedAt = time.Now()
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	_ = e.persistWorkflow(&snapshot)
}

func (e *Engine) RetryFailedStep(ctx context.Context, id string) error {
	if ctx == nil {
		return errors.New("nil context")
	}
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
	snapshot := cloneWorkflow(w)
	e.mu.Unlock()
	if err := e.persistWorkflow(&snapshot); err != nil {
		return err
	}
	return e.ExecuteStep(ctx, id)
}
