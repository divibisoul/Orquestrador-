package prefrontal

import "github.com/divibisoul/Orquestrador-/core/trinity"

func approve(d trinity.Decision, threshold float64) trinity.Decision {
	if threshold <= 0 { threshold = 0.75 }
	if d.ConflictScore >= threshold { d.Approved = false; if d.Reason == "" { d.Reason = "conflict threshold exceeded" }; return d }
	d.Approved = true
	if d.Reason == "" { d.Reason = "approved" }
	return d
}
