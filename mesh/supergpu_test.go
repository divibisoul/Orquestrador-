package mesh

import sg "github.com/divibisoul/Orquestrador-/supergpu"

var supergpu = struct {
	New func(sg.Backend) *sg.Runtime
}{
	New: sg.New,
}
