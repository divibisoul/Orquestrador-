package cognitive

import (
 "context"
 "errors"
 "fmt"
 "strings"
 "time"

 "github.com/divibisoul/Orquestrador-/backend"
 "github.com/divibisoul/Orquestrador-/prefrontal"
)

type Executor interface {
 Execute(context.Context,string,map[string]any,string)(map[string]any,string,error)
}
type Critic struct{ cortex *prefrontal.Cortex; sara *backend.SARAProxy; cfg Config }

func NewCritic(c *prefrontal.Cortex,s *backend.SARAProxy,cfg Config)(*Critic,error){
 if c==nil{return nil,errors.New("prefrontal cortex is required")}
 return &Critic{cortex:c,sara:s,cfg:cfg},nil
}

func(c *Critic) Check(ctx context.Context,g Goal)error{
 candidate:=prefrontal.Candidate{ID:g.ID,Risk:g.Risk,Cost:g.Cost,Urgency:g.Urgency,Impact:g.Impact,Utility:1-g.Risk,Uncertainty:0,Score:1-g.Risk}
 if err:=c.cortex.ValidateAction(candidate);err!=nil{return fmt.Errorf("LOCAL_CRITIQUE_BLOCKED: %w",err)}
 if g.Irreversible && c.cfg.RequireSaraForPolicy {
  if c.sara==nil || !c.sara.Configured(){return errors.New("SARA_POLICY_UNAVAILABLE")}
  if _,err:=c.sara.Audit(ctx,g.Objective,g.CorrelationID);err!=nil{return fmt.Errorf("SARA_POLICY_AUDIT_FAILED: %w",err)}
 }
 return nil
}

type Loop struct{
 cfg Config
 goals *GoalStore
 memory *WorkingMemory
 executor Executor
 planner *Planner
 critic *Critic
 observer func(Observation)
 store *backend.SupabaseStore
}

func New(cfg Config,goals *GoalStore,memory *WorkingMemory,executor Executor,planner *Planner,critic *Critic,store *backend.SupabaseStore)(*Loop,error){
 if goals==nil||memory==nil||executor==nil||planner==nil||critic==nil{return nil,errors.New("goal,memory,executor,planner and critic are required")}
 if cfg.WorkingMemoryItems<=0||cfg.WorkingMemoryTTL<=0||cfg.GoalTTL<=0{return nil,errors.New("invalid cognitive runtime bounds")}
 return &Loop{cfg:cfg,goals:goals,memory:memory,executor:executor,planner:planner,critic:critic,store:store},nil
}

func(l *Loop) Enabled()bool{return l!=nil&&l.cfg.Enabled}

func(l *Loop) Run(ctx context.Context,g Goal)([]Observation,error){
 if !l.Enabled(){return nil,errors.New("COGNITIVE_LOOP_DISABLED")}
 if ctx==nil{return nil,errors.New("context is nil")}
 if strings.TrimSpace(g.ID)==""||strings.TrimSpace(g.Objective)==""{return nil,errors.New("goal id and objective are required")}
 if strings.TrimSpace(g.CorrelationID)==""{return nil,errors.New("goal correlation_id is required")}
 if len(g.Capabilities)==0{return nil,errors.New("goal requires capabilities")}
 if g.ExpiresAt.IsZero(){g.ExpiresAt=time.Now().Add(l.cfg.GoalTTL)}
 if err:=l.goals.Put(g);err!=nil{return nil,err}
 if err:=l.critic.Check(ctx,g);err!=nil{return nil,err}
 out:=make([]Observation,0,len(g.Capabilities))
 for i,capability:=range g.Capabilities{
  capability=strings.TrimSpace(capability); if capability==""{return out,errors.New("empty capability")}
  step:=Step{ID:fmt.Sprintf("%s-step-%d",g.ID,i+1),GoalID:g.ID,Capability:capability,Payload:cloneMap(g.Input),CorrelationID:g.CorrelationID}
  started:=time.Now()
  value,peer,err:=l.executor.Execute(ctx,step.Capability,step.Payload,step.CorrelationID)
  obs:=Observation{GoalID:g.ID,StepID:step.ID,Capability:step.Capability,OK:err==nil,Peer:peer,LatencyMS:time.Since(started).Milliseconds(),CorrelationID:g.CorrelationID,At:time.Now().UTC()}
  if err!=nil{obs.Error=err.Error()}
  out=append(out,obs)
  if l.observer!=nil{l.observer(obs)}
  _=l.memory.Put(step.ID,map[string]any{"output":value,"observation":obs},1)
  l.persist(ctx,obs)
  if err!=nil{return out,fmt.Errorf("goal step %s failed: %w",step.ID,err)}
 }
 return out,nil
}

func(l *Loop) SetObserver(fn func(Observation)){l.observer=fn}

func(l *Loop) persist(ctx context.Context,o Observation){
 if l.store==nil||!l.store.Configured(){return}
 status:="error";if o.OK{status="ok"}
 _=l.store.RecordRun(ctx,map[string]any{"trace_id":o.CorrelationID,"correlation_id":o.CorrelationID,"source":"N07.cognitive","status":status,"metadata":map[string]any{"goal_id":o.GoalID,"step_id":o.StepID,"capability":o.Capability,"peer":o.Peer,"latency_ms":o.LatencyMS,"error":o.Error}})
}
