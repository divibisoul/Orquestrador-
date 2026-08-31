package transcendental

import (
	"context"
	"testing"
	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func TestSelectorHeuristics(t *testing.T){cases:=[]struct{w trinity.Workload;want string}{{trinity.Workload{Precision:"fp64"},"blackwell"},{trinity.Workload{MemoryNeeded:500},"mi400"},{trinity.Workload{MatrixSize:5000},"vera_rubin"},{trinity.Workload{MatrixSize:500},"trillium"},{trinity.Workload{BatchSize:100},"blackwell"}};for _,c:=range cases{if got:=Select(c.w).Name;got!=c.want{t.Fatalf("got %s want %s",got,c.want)}}}
func TestComputeEngine(t *testing.T){e:=NewEngine(trinity.ComputeConfig{});r,err:=e.Execute(context.Background(),trinity.Workload{ID:"x",Kind:"text"},trinity.Route{Model:"blackwell"});if err!=nil||!r.Success{t.Fatalf("result=%+v err=%v",r,err)}}
