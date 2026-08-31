package tests

import (
	"testing"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/selection"
)

func TestSelectorSmallPrefersTrillium(t *testing.T) {
	cfg:=core.DefaultConfig(); wl:=core.Workload{ID:"small",Operation:"matmul",Precision:core.BF16,MatrixSize:512,BatchSize:1,DataBytes:1<<20,MemoryNeeded:1}
	s,err:=selection.Select(wl,models.DefaultCatalog(cfg),"auto",cfg.PrecisionFallback);if err!=nil{t.Fatal(err)};if s.Model.Name()!="Google TPU v6e Trillium"{t.Fatalf("got %s",s.Model.Name())}
}
func TestSelectorLargePrefersHighBandwidth(t *testing.T) {
	cfg:=core.DefaultConfig(); wl:=core.Workload{ID:"large",Operation:"matmul",Precision:core.FP8,MatrixSize:8192,BatchSize:4,DataBytes:1<<32,MemoryNeeded:300}
	s,err:=selection.Select(wl,models.DefaultCatalog(cfg),"auto",cfg.PrecisionFallback);if err!=nil{t.Fatal(err)};if s.Model.GetMemoryCapacityGB()<300{t.Fatalf("selected insufficient memory model %s",s.Model.Name())}
}
func TestSelectorFP64RestrictsFamily(t *testing.T) {
	cfg:=core.DefaultConfig(); wl:=core.Workload{ID:"fp64",Operation:"matmul",Precision:core.FP64,MatrixSize:2048,BatchSize:1,DataBytes:1<<20,MemoryNeeded:1}
	s,err:=selection.Select(wl,models.DefaultCatalog(cfg),"auto",cfg.PrecisionFallback);if err!=nil{t.Fatal(err)};if s.EffectivePrecision!=core.FP64{t.Fatalf("precision fallback unexpected: %s",s.EffectivePrecision)}
}
