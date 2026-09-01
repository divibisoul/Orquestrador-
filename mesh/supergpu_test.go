package mesh

import sg "github.com/divibisoul/Orquestrador-/supergpu"

var supergpuCompat = struct {
	New func(sg.Backend) *sg.Runtime
}{
	New: sg.New,
}
