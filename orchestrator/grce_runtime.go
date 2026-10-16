package orchestrator

// GRCE runtime implementation lives in backend/grce_runtime.go because the
// runtime binds SARAProxy and the Nervo Vago gateway. Keeping that adapter out
// of this package avoids the existing backend -> orchestrator dependency cycle.
// The GRCE operation registration boundary remains here in grce_ops.go.
