package neuralfabric

import (
	"context"
	"sync"
	"testing"
	"github.com/divibisoul/Orquestrador-/core/trinity"
)

func TestFabricRoute(t *testing.T){f:=NewFabric(trinity.FabricConfig{}); r,err:=f.Route(context.Background(),trinity.Strategy{Precision:"fp32"},trinity.Workload{ID:"x",Kind:"matrix",MatrixSize:5000}); if err!=nil||r.Model==""{t.Fatalf("route=%+v err=%v",r,err)}}
func TestFabricFeedback(t *testing.T){f:=NewFabric(trinity.FabricConfig{}); r,_:=f.Route(context.Background(),trinity.Strategy{},trinity.Workload{ID:"x",Kind:"text"}); if err:=f.Feedback(context.Background(),trinity.Feedback{WorkloadID:"x",Route:r,Success:true,Quality:1});err!=nil{t.Fatal(err)}}
func TestConcurrency(t *testing.T){f:=NewFabric(trinity.FabricConfig{}); w:=trinity.Workload{ID:"x",Kind:"text"}; var wg sync.WaitGroup; for i:=0;i<50;i++{wg.Add(1);go func(){defer wg.Done();_,_=f.Route(context.Background(),trinity.Strategy{},w);_=f.Feedback(context.Background(),trinity.Feedback{WorkloadID:"x",Route:trinity.Route{Model:"blackwell"},Success:true})}()};wg.Wait()}
