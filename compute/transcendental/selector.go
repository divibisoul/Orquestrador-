package transcendental

import "github.com/divibisoul/Orquestrador-/core/trinity"

func Select(w trinity.Workload) Model {
	name := "blackwell"
	switch {
	case w.Precision == "fp64": name="blackwell"
	case w.MemoryNeeded > 400: name="mi400"
	case w.MatrixSize > 4096: name="vera_rubin"
	case w.MatrixSize > 0 && w.MatrixSize < 1024: name="trillium"
	case w.BatchSize > 64: name="blackwell"
	}
	for _, m := range Models { if m.Name==name { return m } }
	return Models[0]
}
