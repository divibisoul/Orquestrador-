package tests

import (
	"context"
	"testing"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/core"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/executor"
	"github.com/divibisoul/Orquestrador-/compute/transcendental/models"
)

func TestExecutorDisabledIsNonOperational(t *testing.T) {
	cfg:=core.DefaultConfig(); eng,err:=core.NewEngine(cfg);if err!=nil{t.Fatal(err)}; ex,err:=executor.New(eng,models.DefaultCatalog(cfg),"auto");if err!=nil{t.Fatal(err)}
	_,err=ex.Execute(context.Background(),core.Workload{ID:"disabled",Operation:"matmul",Precision:core.FP8,MatrixSize:1024,BatchSize:1});if err==nil{t.Fatal("expected disabled error")}
}

func TestExecutorDeterministicAndMetrics(t *testing.T) {
	cfg:=core.DefaultConfig();cfg.Enabled=true;eng,err:=core.NewEngine(cfg);if err!=nil{t.Fatal(err)};ex,err:=executor.New(eng,models.DefaultCatalog(cfg),"auto");if err!=nil{t.Fatal(err)}
	wl:=core.Workload{ID:"det",Operation:"matmul",Precision:core.FP8,MatrixSize:4096,BatchSize:2,DataBytes:1<<28,MemoryNeeded:64}
	a,err:=ex.Estimate(context.Background(),wl);if err!=nil{t.Fatal(err)};b,err:=ex.Estimate(context.Background(),wl);if err!=nil{t.Fatal(err)};if a!=b{t.Fatalf("estimates are not deterministic: %#v != %#v",a,b)}
	r,err:=ex.Execute(context.Background(),wl);if err!=nil{t.Fatal(err)};if string(r.Data)!="simulated-result"{t.Fatalf("unexpected result data %q",r.Data)};if r.Metrics.Architecture==""||r.Metrics.LatencyMs<=0{t.Fatalf("invalid metrics %#v",r.Metrics)}
}
