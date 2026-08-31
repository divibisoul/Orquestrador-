package tests

import (
	"testing"
	"time"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
)

func TestReferenceModels(t *testing.T) {
	cfg:=core.DefaultConfig(); cat:=models.DefaultCatalog(cfg); if len(cat)!=5{t.Fatalf("models=%d want 5",len(cat))}
	checks:=[]struct{name string;p core.Precision;want float64}{
		{"NVIDIA Blackwell B200",core.FP4,9},
		{"NVIDIA Vera Rubin (preliminary 2026)",core.FP4,50},
		{"AMD Instinct MI400 / CDNA5",core.FP4,40},
		{"Google TPU v6e Trillium",core.BF16,.918},
		{"Huawei Atlas 950 (scale-out reference)",core.FP4,20},
	}
	for _,c:=range checks { var found models.PerformanceModel; for _,m:=range cat{if m.Name()==c.name{found=m}};if found==nil{t.Fatalf("missing %s",c.name)};if got:=found.GetPFLOPS(c.p);got!=c.want{t.Fatalf("%s PFLOPS=%v want %v",c.name,got,c.want)};if found.EstimateTime(core.Workload{ID:"w",Operation:"matmul",Precision:c.p,MatrixSize:2048,BatchSize:1,DataBytes:1<<20,MemoryNeeded:1})<=0{t.Fatalf("%s non-positive estimate",c.name)} }
	_ = time.Second
}
