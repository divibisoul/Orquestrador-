package orchestrator

import "time"

func (f *Federation) Stats() map[string]any {
	f.mu.RLock()
	defer f.mu.RUnlock()

	healthy, calls, success, open := 0, 0, 0, 0
	now := time.Now()
	details := make([]map[string]any, 0, len(f.peers))
	for _, state := range f.peers {
		if state.Healthy {
			healthy++
		}
		calls += int(state.calls)
		success += int(state.success)
		circuitOpen := now.Before(state.openUntil)
		if circuitOpen {
			open++
		}
		details = append(details, map[string]any{
			"nucleus":      state.Nucleus,
			"healthy":      state.Healthy,
			"latency_ms":   float64(state.Latency) / float64(time.Millisecond),
			"success_rate": state.SuccessRate,
			"in_flight":    state.InFlight,
			"calls":        state.calls,
			"success":      state.success,
			"circuit_open": circuitOpen,
		})
	}
	globalSuccessRate := float64(0)
	if calls > 0 {
		globalSuccessRate = float64(success) / float64(calls)
	}
	return map[string]any{
		"peers":         len(f.peers),
		"healthy":       healthy,
		"calls":         calls,
		"success":       success,
		"success_rate":  globalSuccessRate,
		"circuit_open":  open,
		"details":       details,
	}
}
