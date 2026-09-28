package cognitive

import (
 "context"
 "errors"
 "fmt"
 "strings"
 "time"
)

func (l *Loop) RunPlanned(ctx context.Context,g Goal)([]Observation,error){
 if !l.Enabled(){return nil,errors.New("COGNITIVE_LOOP_DISABLED")}
 if ctx==nil{return nil,errors.New("context is nil")}
 if strings.TrimSpace(g.ID)==""||strings.TrimSpace(g.Objective)==""{return nil,errors.New("goal id and objective are required")}
 if strings.TrimSpace(g.CorrelationID)==""{return nil,errors.New("goal correlation_id is required")}
 if g.ExpiresAt.IsZero(){g.ExpiresAt=time.Now().Add(l.cfg.GoalTTL)}
 if !g.ExpiresAt.After(time.Now().UTC()){return nil,errors.New("goal expired")}
 if err:=l.goals.Put(g);err!=nil{return nil,err}
 if err:=l.critic.Check(ctx,g);err!=nil{return nil,err}
 steps,err:=l.planner.Plan(ctx,g)
 if err!=nil{return nil,err}
 out:=make([]Observation,0,len(steps))
 batcher,ok:=l.executor.(*OctaCoreExecutor)
 if !ok{return nil,errors.New("cognitive executor is not the canonical Octacore executor")}
 observations:=batcher.ExecuteSteps(ctx,steps)
 if len(observations)!=len(steps){return nil,errors.New("COGNITIVE_EXECUTOR_RESULT_CARDINALITY_MISMATCH")}
 anyFailure:=false
 for _,obs:=range observations{
  out=append(out,obs)
  if l.observer!=nil{l.observer(obs)}
  relevance:=1.0
  if !obs.OK{relevance=0.25;anyFailure=true}
  if err:=l.memory.Put(obs.StepID,map[string]any{"output":obs.Output,"observation":obs},relevance);err!=nil{return out,fmt.Errorf("WORKING_MEMORY_PERSISTENCE_FAILED:%w",err)}
  if err:=l.persist(ctx,obs);err!=nil{return out,err}
 }
 if anyFailure{return out,errors.New("COGNITIVE_STEP_FAILED")}

 return out,nil
}

func NewPlannerFromExecutor(e Executor)(*Planner,error){
 switch x:=e.(type){
 case *MeshExecutor:
  return NewPlanner(x.peers)
 default:
  return nil,errors.New("planner requires the canonical MeshExecutor")
 }
}
