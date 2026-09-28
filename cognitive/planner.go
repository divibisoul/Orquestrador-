package cognitive

import (
 "context"
 "errors"
 "fmt"
 "strings"
 "github.com/divibisoul/Orquestrador-/mesh"
)

type ToolDescriptor struct {
 Capability string `json:"capability"`
 Owner string `json:"owner"`
 Executable bool `json:"executable"`
 InputSchema map[string]any `json:"input_schema,omitempty"`
 OutputSchema map[string]any `json:"output_schema,omitempty"`
}

type Planner struct{ peers *mesh.PeerClient }

func NewPlanner(peers *mesh.PeerClient)(*Planner,error){
 if peers==nil{return nil,errors.New("mesh peer client is required")}
 return &Planner{peers:peers},nil
}

func(p *Planner) DiscoverTool(ctx context.Context,capability string)(ToolDescriptor,error){
 capability=strings.TrimSpace(capability);if capability==""{return ToolDescriptor{},errors.New("capability is required")}
 for _,peer:=range p.peers.ConfiguredPeers(){
  description,err:=p.peers.Discover(ctx,peer.Nucleus);if err!=nil{continue}
  if executable(description,capability){
   return ToolDescriptor{Capability:capability,Owner:peer.Nucleus,Executable:true,InputSchema:schema(description,capability,"input"),OutputSchema:schema(description,capability,"output")},nil
  }
 }
 return ToolDescriptor{},fmt.Errorf("TOOL_NOT_DISCOVERED:%s",capability)
}

func(p *Planner) Plan(ctx context.Context,g Goal)([]Step,error){
 if ctx==nil{return nil,errors.New("context is nil")}
 if len(g.Capabilities)==0{return nil,errors.New("goal requires capabilities")}
 if len(g.Capabilities) > 8 {return nil,errors.New("goal exceeds cognitive parallel step limit")}
 out:=make([]Step,0,len(g.Capabilities))
 for i,capability:=range g.Capabilities{
  tool,err:=p.DiscoverTool(ctx,capability);if err!=nil{return nil,err}
  group:=fmt.Sprintf("%s-pre",g.ID)
  out=append(out,Step{ID:fmt.Sprintf("%s-step-%d",g.ID,i+1),GoalID:g.ID,Capability:tool.Capability,Target:tool.Owner,BackendPrefs:[]string{"REMOTE_MESH"},Kind:classifyKind(capability),Payload:cloneMap(g.Input),ParallelGroup:&group,CorrelationID:g.CorrelationID})
 }
 return out,nil
}

func executable(d map[string]any,capability string)bool{
 raw:=d["executableCapabilities"];if raw==nil{if p,ok:=d["payload"].(map[string]any);ok{raw=p["executableCapabilities"]}}
 switch x:=raw.(type){
 case []any:for _,v:=range x{if s,ok:=v.(string);ok&&sameCapability(s,capability){return true}}
 case []string:for _,s:=range x{if sameCapability(s,capability){return true}}
 }
 return false
}

func sameCapability(a,b string)bool{
 an,av:=splitCapability(a);bn,bv:=splitCapability(b);return an==bn&&(bv==""||av==bv)
}
func splitCapability(v string)(string,string){
 parts:=strings.SplitN(strings.TrimSpace(v),"@",2);if len(parts)==1{return parts[0],""};return parts[0],parts[1]
}
func schema(d map[string]any,capability,side string)map[string]any{
 raw,ok:=d["capabilitySchemas"].(map[string]any);if !ok{return nil}
 item,ok:=raw[capability].(map[string]any);if !ok{return nil}
 v,_:=item[side].(map[string]any);return v
}


func classifyKind(capability string) string {
 switch {
 case strings.HasPrefix(capability, "research."): return "research"
 case strings.HasPrefix(capability, "perceive."), strings.HasPrefix(capability, "audio."): return "perceive"
 case strings.HasPrefix(capability, "tool:"), strings.HasPrefix(capability, "tool."): return "tool"
 case strings.HasPrefix(capability, "session."), strings.HasPrefix(capability, "support."): return "session_step"
 default: return "custom"
 }
}
