package cognitive

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "strings"

 "github.com/divibisoul/Orquestrador-/orchestrator"
 "github.com/divibisoul/Orquestrador-/protocol"
)

const OpGoalRun = "cognitive.goal.run@1.0.0"
const OpHealth = "cognitive.health@1.0.0"

func RegisterOperations(e *orchestrator.Engine,l *Loop) error{
 if e==nil{return errors.New("orchestrator engine is required")}
 if l==nil{return errors.New("cognitive loop is required")}
 if !l.Enabled(){return errors.New("COGNITIVE_LOOP_DISABLED")}
 if err:=e.Register(OpGoalRun,func(ctx context.Context,m protocol.Message)(protocol.Result,error){
  raw:=strings.TrimSpace(m.Metadata["cognitive_goal_json"]);if raw==""{return result(m,nil,errors.New("metadata.cognitive_goal_json is required"))}
  var g Goal;if err:=json.Unmarshal([]byte(raw),&g);err!=nil{return result(m,nil,fmt.Errorf("INVALID_COGNITIVE_GOAL: %w",err))}
  if g.CorrelationID==""{g.CorrelationID=m.CorrelationID}
  obs,err:=l.RunPlanned(ctx,g);rawOut,_:=json.Marshal(map[string]any{"goal":g,"observations":obs})
  return result(m,rawOut,err)
 });err!=nil{return err}
 return e.Register(OpHealth,func(_ context.Context,m protocol.Message)(protocol.Result,error){
  status := "DISABLED"
  if l.Enabled() { status = "READY" }
  raw,_:=json.Marshal(map[string]any{"status":status,"enabled":l.Enabled(),"working_memory_items":len(l.memory.Snapshot()),"octacore":true})
  return result(m,raw,nil)
 })
}

func result(m protocol.Message,raw []byte,err error)(protocol.Result,error){
 r:=protocol.Result{TraceID:m.TraceID,CorrelationID:m.CorrelationID,Source:"N07.cognitive",Target:m.Source,Status:"ok",Metadata:map[string]string{"cognitive_json":string(raw)}}
 if err!=nil{r.Status="error";r.Error=err.Error()}
 return r,err
}


func ConfigFromEnv() Config { return cognitiveConfigFromEnv() }
