package tests

import("context";"testing";"github.com/divibisoul/Orquestrador-/core/orchestrator/execution";"github.com/divibisoul/Orquestrador-/core/orchestrator/workflow")
func TestWorkflowLifecycle(t *testing.T){e:=workflow.NewEngine();id,err:=e.CreateWorkflow(workflow.WorkflowSpec{Goal:"test",Steps:[]string{"a","b"}});if err!=nil{t.Fatal(err)};if err=e.ExecuteStep(id);err!=nil{t.Fatal(err)};if e.GetWorkflowStatus(id)!=workflow.Running{t.Fatal("expected running")};if err=e.ExecuteStep(id);err!=nil{t.Fatal(err)};if e.GetWorkflowStatus(id)!=workflow.Completed{t.Fatal("expected completed")}}
func TestParallel(t *testing.T){e:=execution.New(2);errs:=e.ExecuteParallel(context.Background(),[]execution.Task{func(context.Context)error{return nil},func(context.Context)error{return nil}});if len(errs)!=0{t.Fatal(errs)}}
func TestRateLimiter(t *testing.T){e:=execution.New(1);if !e.RateLimiter("x",1){t.Fatal("first token denied")};if e.RateLimiter("x",1){t.Fatal("second token unexpectedly allowed")}}
