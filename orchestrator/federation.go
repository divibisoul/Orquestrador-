package orchestrator

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "sort"
    "strings"
)

type PeerInvoker interface {
    CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
    CallBest(context.Context, string, map[string]any, string) (map[string]any, string, error)
}

type FederatedCapability struct {
    ID string `json:"id"`
    Owner string `json:"owner"`
    Kind string `json:"kind"`
    Composable bool `json:"composable"`
    Consumers []string `json:"consumers"`
    Tool string `json:"tool,omitempty"`
}

type Federation struct { peer PeerInvoker; catalog map[string]FederatedCapability }

func NewFederation(peer PeerInvoker) *Federation {
    f := &Federation{peer: peer, catalog: make(map[string]FederatedCapability)}
    f.loadCatalog()
    return f
}

func (f *Federation) loadCatalog() {
    n01 := []string{"android.device_info","android.battery","android.memory","android.network","android.events","shizuku.bridge","brightness","wifi","bluetooth","airplane","background-process"}
    for _, id := range n01 { f.catalog[id] = FederatedCapability{ID:id,Owner:"N01",Kind:"native",Composable:true,Consumers:[]string{"N07"}} }
    n06 := []string{"support.context","support.artifacts","support.documents","support.tool-execution","support.streaming","support.mesh","support.ai-pilot"}
    for _, id := range n06 { f.catalog[id] = FederatedCapability{ID:id,Owner:"N06",Kind:"capability",Composable:true,Consumers:[]string{"N07"}} }
    for _, tool := range []string{"createDocument","updateDocument","getWeather","requestSuggestions"} {
        id := "tool:" + tool
        f.catalog[id] = FederatedCapability{ID:id,Owner:"N06",Kind:"tool",Tool:tool,Composable:true,Consumers:[]string{"N07"}}
    }
}

func (f *Federation) Catalog() []FederatedCapability {
    out := make([]FederatedCapability,0,len(f.catalog))
    for _, c := range f.catalog { c.Consumers = append([]string(nil), c.Consumers...); out = append(out,c) }
    sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
    return out
}

func (f *Federation) Resolve(id string) (FederatedCapability,error) {
    id = strings.TrimSpace(id)
    c,ok := f.catalog[id]
    if !ok { return FederatedCapability{}, fmt.Errorf("federated capability not registered: %s",id) }
    return c,nil
}

func (f *Federation) Invoke(ctx context.Context,id string,payload map[string]any,correlation string) (map[string]any,error) {
    if f.peer == nil { return nil, errors.New("peer federation is not configured") }
    c,err := f.Resolve(id); if err != nil { return nil,err }
    if correlation=="" { return nil,errors.New("correlation is required") }
    envelope := map[string]any{"federated":true,"capability":id,"owner":c.Owner,"payload":payload}
    if c.Kind=="tool" { envelope["tool"]=c.Tool }
    return f.peer.CallWithCorrelation(ctx,c.Owner,id,envelope,correlation)
}

func (f *Federation) Compose(ctx context.Context,ids []string,payload map[string]any,correlation string) (map[string]any,error) {
    if len(ids)<2 { return nil,errors.New("composition requires at least two capabilities") }
    current:=payload
    components:=make([]string,0,len(ids))
    for _,id := range ids {
        c,err:=f.Resolve(id); if err!=nil{return nil,err}
        if !c.Composable{return nil,fmt.Errorf("capability is not composable: %s",id)}
        components=append(components,id)
        result,err:=f.Invoke(ctx,id,current,correlation);if err!=nil{return nil,err}
        current=map[string]any{"input":current,"result":result}
    }
    return map[string]any{"type":"composed","components":components,"result":current,"correlationId":correlation},nil
}

func DecodeData(raw string) (map[string]any,error) {
    if strings.TrimSpace(raw)=="" { return map[string]any{},nil }
    var data map[string]any
    if err:=json.Unmarshal([]byte(raw),&data);err!=nil{return nil,err}
    return data,nil
}
