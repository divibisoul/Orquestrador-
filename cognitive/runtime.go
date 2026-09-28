package cognitive

import (
 "errors"
 "github.com/divibisoul/Orquestrador-/backend"
 "github.com/divibisoul/Orquestrador-/mesh"
 "github.com/divibisoul/Orquestrador-/prefrontal"
 "github.com/divibisoul/Orquestrador-/octacore"
)

func Build(cfg Config,cortex *prefrontal.Cortex,peers *mesh.PeerClient,processor *octacore.Processor,sara *backend.SARAProxy,store *backend.SupabaseStore)(*Loop,error){
 if !cfg.Enabled{return nil,errors.New("COGNITIVE_LOOP_DISABLED")}
 if peers==nil{return nil,errors.New("mesh peer client is required")}
 executor,err:=NewOctaCoreExecutor(processor,peers);if err!=nil{return nil,err}
 planner,err:=NewPlanner(peers);if err!=nil{return nil,err}
 critic,err:=NewCritic(cortex,sara,cfg);if err!=nil{return nil,err}
 return New(cfg,NewGoalStore(),NewWorkingMemory(cfg),executor,planner,critic,store)
}
