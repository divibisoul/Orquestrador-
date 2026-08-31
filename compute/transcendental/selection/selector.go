package selection

import (
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
)

type Selection struct {
	Model               models.PerformanceModel
	EffectivePrecision  core.Precision
	FallbackUsed        bool
	Penalty             float64
}

func Select(wl core.Workload, catalog []models.PerformanceModel, mode string, strategy string, fallback core.Precision) (Selection, error) {
	if len(catalog)==0 { return Selection{}, errors.New("no performance models available") }
	if wl.MatrixSize<1 || wl.BatchSize<1 { return Selection{}, errors.New("matrix size and batch size must be >= 1") }
	if fallback=="" { fallback=core.FP32 }
	candidates:=make([]Selection,0,len(catalog))
	for _, model := range catalog {
		if model==nil || model.GetMemoryCapacityGB()<wl.MemoryNeeded { continue }
		if mode!="" && mode!="auto" && !matchesMode(model.Name(),mode) { continue }
		precision:=wl.Precision; fallbackUsed:=false; penalty:=0.0
		if wl.Precision==core.FP64 && !isFP64Family(model.Name()) { continue }
		if !model.Supports(precision) {
			if model.Supports(fallback) { precision=fallback; fallbackUsed=true; penalty=.25 } else { continue }
		}
		candidates=append(candidates,Selection{Model:model,EffectivePrecision:precision,FallbackUsed:fallbackUsed,Penalty:penalty})
	}
	if len(candidates)==0 { return Selection{}, errors.New("no compatible model satisfies mode, memory, and precision requirements") }
	best:=candidates[0]; bestScore:=score(wl,best,strategy)
	for _,candidate:=range candidates[1:] { s:=score(wl,candidate,strategy);if s<bestScore||(s==bestScore&&candidate.Model.Name()<best.Model.Name()){best,bestScore=candidate,s} }
	return best,nil
}

func matchesMode(name, mode string) bool {
	n:=strings.ToLower(name); switch strings.ToLower(mode) {
	case "blackwell": return strings.Contains(n,"blackwell")
	case "vera_rubin": return strings.Contains(n,"rubin")
	case "cdna5": return strings.Contains(n,"cdna5")
	case "trillium": return strings.Contains(n,"trillium")
	case "atlas": return strings.Contains(n,"atlas 950")
	default: return false
	}
}
func score(wl core.Workload,s Selection,strategy string) float64 {
	t:=float64(s.Model.EstimateTime(core.Workload{ID:wl.ID,Operation:wl.Operation,Precision:s.EffectivePrecision,MatrixSize:wl.MatrixSize,BatchSize:wl.BatchSize,DataBytes:wl.DataBytes,MemoryNeeded:wl.MemoryNeeded,Priority:wl.Priority,Metadata:wl.Metadata}));mem:=float64(max64(wl.MemoryNeeded,0));base:=t.Seconds()
	if strings.EqualFold(strategy,"latency_first"){return base*(1+s.Penalty)+mem*.000001}
	if strings.EqualFold(strategy,"memory_first"){return mem+base*.001*(1+s.Penalty)}
	if wl.MatrixSize<1024&&strings.Contains(strings.ToLower(s.Model.Name()),"trillium"){return base*.5+s.Penalty}
	if wl.MatrixSize>4096&&(s.Model.GetBandwidthGBs()>10000||s.Model.GetMemoryCapacityGB()>200){return base*.8+s.Penalty}
	return base*(1+s.Penalty)+mem*.000001
}
func isFP64Family(name string) bool { n:=strings.ToLower(name);return strings.Contains(n,"blackwell")||strings.Contains(n,"rubin") }
func max64(a,b int64) int64 {if a>b{return a};return b}
