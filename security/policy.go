package security

import "errors"

type Policy struct {
	AllowExternalEffects bool
	MaxCost              float64
	MinConfidence       float64
}

// Validate enforces the current execution policy. Authentication and
// authorization are intentionally separate concerns and are not implied by
// this function.
func Validate(p Policy, cost, confidence float64) error {
	if cost < 0 {
		return errors.New("invalid cost")
	}
	if confidence < 0 || confidence > 1 {
		return errors.New("invalid confidence")
	}
	if p.MaxCost >= 0 && cost > p.MaxCost {
		return errors.New("cost exceeds policy")
	}
	if confidence < p.MinConfidence {
		return errors.New("confidence below policy")
	}
	if !p.AllowExternalEffects {
		return errors.New("external effects disabled")
	}
	return nil
}

// Authorize is a deliberately small extension point for future identity-aware
// authorization. Callers must supply an authenticated principal; an empty
// principal is rejected rather than treated as anonymous access.
func Authorize(principal, action string) error {
	if principal == "" {
		return errors.New("authentication required")
	}
	if action == "" {
		return errors.New("action required")
	}
	return nil
}

// AuditRecord is the minimum immutable audit payload expected at the security
// boundary. Persistence is intentionally left to the runtime integration.
type AuditRecord struct {
	Principal string
	Action    string
	Allowed   bool
	Reason    string
}
