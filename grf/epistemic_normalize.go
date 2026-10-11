package grf

import (
	"fmt"
	"strings"
)

// NormalizeEpistemic is the non-erroring compatibility helper. Unknown labels
// normalize to UNRESOLVED, which is never an activation state.
func NormalizeEpistemic(raw string) EpistemicState {
	normalized, err := NormalizeEpistemicStrict(raw)
	if err != nil {
		return UNRESOLVED
	}
	return normalized
}

// NormalizeEpistemicStrict normalizes case and whitespace while rejecting
// unknown labels so boundary adapters can fail closed instead of guessing.
func NormalizeEpistemicStrict(raw string) (EpistemicState, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(REAL):
		return REAL, nil
	case string(PROJECTED):
		return PROJECTED, nil
	case string(BLOCKED):
		return BLOCKED, nil
	case string(UNRESOLVED):
		return UNRESOLVED, nil
	case string(PRESERVED):
		return PRESERVED, nil
	case string(ACTIVE):
		return ACTIVE, nil
	case string(IDLE):
		return IDLE, nil
	default:
		return UNRESOLVED, fmt.Errorf("GRF_EPISTEMIC_UNKNOWN:%q", raw)
	}
}
