package mesh

import(
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"github.com/divibisoul/Orquestrador-/orchestrator"
)

type EnhancedFederatedGateway struct{base *FederatedGateway}
func NewEnhancedFederatedHTTPGateway(engine *orchestrator.Engine)*EnhancedFederatedGateway{return &EnhancedFederatedGateway{base:NewFederatedHTTPGateway(engine)}}
func(g *EnhancedFederatedGateway)ServeHTTP(w http.ResponseWriter,r *http.Request){
	if g==nil||g.base==nil{writeMeshJSON(w,http.StatusServiceUnavailable,map[string]any{"error":"N07 gateway unavailable"});return}
	if r.Method!=http.MethodPost||g.base.engine==nil{g.base.ServeHTTP(w,r);return}
	body,err:=ioReadLimited(r.Body,1<<20);if err!=nil{writeMeshJSON(w,http.StatusBadRequest,map[string]any{"error":err.Error()});return}
	var wire canonicalWireEnvelope;if err:=json.Unmarshal(body,&wire);err!=nil{r.Body=io.NopCloser(strings.NewReader(string(body)));g.base.ServeHTTP(w,r);return}
	capability:=canonicalCapability(wire);if capability!="mesh.supergpu.parallel"&&capability!="supergpu.parallel"{r.Body=io.NopCloser(strings.NewReader(string(body)));g.base.ServeHTTP(w,r);return}
	envelope:=normalizedMeshEnvelope(wire);if err:=envelope.Validate();err!=nil{writeMeshJSON(w,http.StatusBadRequest,gatewayError(envelope,err.Error()));return};if err:=g.base.base.authenticateWire(wire,r,envelope);err!=nil{g.base.base.respond(w,http.StatusUnauthorized,envelope,"ERROR",map[string]any{"error":"mesh authentication failed"});return}
	payload:=envelope.NestedPayload();raw,ok:=payload["tasks"].([]any);if !ok||len(raw)==0{g.base.base.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"supergpu.parallel requires payload.tasks"});return}
	type task struct{ID string;Capability string;Payload map[string]any;Required bool};tasks:=make([]task,0,len(raw));seen:=map[string]struct{}{}
	for i,item:=range raw{m,ok:=item.(map[string]any);if !ok{g.base.base.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"task must be object","index":i});return};capability,_:=m["capability"].(string);capability=strings.TrimSpace(capability);if capability==""{g.base.base.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"task capability is required","index":i});return};id,_:=m["id"].(string);id=strings.TrimSpace(id);if id==""{id=fmt.Sprintf("task-%d",i)};if _,exists:=seen[id];exists{g.base.base.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"duplicate task id","id":id});return};seen[id]=struct{}{};p,_:=m["payload"].(map[string]any);if p==nil{p=map[string]any{}};required,_:=m["required"].(bool);tasks=append(tasks,task{id,capability,p,required})}
	type result struct{ID string `json:"id"`;Capability string `json:"capability"`;Owner string `json:"owner,omitempty"`;Status string `json:"status"`;DurationMs int64 `json:"duration_ms"`;Payload any `json:"payload,omitempty"`;Error string `json:"error,omitempty"`;CorrelationID string `json:"correlationId"`}
	results:=make([]result,len(tasks));var wg sync.WaitGroup;for i,item:=range tasks{i,item=i,item;wg.Add(1);go func(){defer wg.Done();started:=time.Now();child:=envelope.CorrelationID+"/"+item.ID;remote,owner,e:=g.base.peers.CallBest(r.Context(),item.Capability,item.Payload,child);entry:=result{ID:item.ID,Capability:item.Capability,Status:"error",DurationMs:time.Since(started).Milliseconds(),CorrelationID:child};if e!=nil{entry.Error=e.Error()}else{entry.Status="ok";entry.Owner=owner;entry.Payload=remote["payload"]};results[i]=entry}()};wg.Wait();requiredFailed:=false;for i,v:=range results{if v.Status!="ok"&&tasks[i].Required{requiredFailed=true}};status:=http.StatusOK;if requiredFailed{status=http.StatusBadGateway};g.base.base.respond(w,status,envelope,"TASK_RESULT",map[string]any{"execution":"federated-supergpu-parallel","parentCorrelationId":envelope.CorrelationID,"taskCount":len(results),"requiredFailure":requiredFailed,"tasks":results})
}
