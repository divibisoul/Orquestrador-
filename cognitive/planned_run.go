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
 if err:=l.goals.Put(g);err!=nil{return nil,err}
 if err:=l.critic.Check(ctx,g);err!=nil{return nil,err}
 planner,err:=NewPlannerFromExecutor(l.executor)
 if err!=nil{return nil,err}
 steps,err:=planner.Plan(ctx,g)
 if err!=nil{return nil,err}
 out:=make([]Observation,0,len(steps))
 for _,step:=range steps{
  started:=time.Now()
  value,peer,execErr:=l.executor.Execute(ctx,step.Capability,step.Payload,step.CorrelationID)
  obs:=Observation{GoalID:g.ID,StepID:step.ID,Capability:step.Capability,OK:execErr==nil,Peer:peer,LatencyMS:time.Since(started).Milliseconds(),CorrelationID:g.CorrelationID,At:time.Now().UTC()}
  if execErr!=nil{obs.Error=execErr.Error()}
  out=append(out,obs)
  if l.observer!=nil{l.observer(obs)}
  _=l.memory.Put(step.ID,map[string]any{"output":value,"observation":obs},1)
  l.persist(ctx,obs)
  if execErr!=nil{return out,fmt.Errorf("goal step %s failed: %w",step.ID,execErr)}
 }
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
