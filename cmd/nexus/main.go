package main

import (
 "fmt"
 "github.com/divibisoul/Orquestrador-/core/orchestrator/workflow"
)

func main() {
 engine := workflow.NewEngine()
 id, err := engine.CreateWorkflow(workflow.WorkflowSpec{Goal:"bootstrap", Steps:[]string{"validate","ready"}})
 if err != nil { panic(err) }
 fmt.Printf("orchestrator-nexus workflow=%s status=%s\n", id, engine.GetWorkflowStatus(id))
}
