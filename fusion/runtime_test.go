package fusion

import (
	"context"
	"testing"
)

func TestAdjacentDynamicFusionExecutesRegisteredComponents(t *testing.T) {
	r := NewRegistry()
	mk := func(id,nucleus string, add float64, kind Kind) Component { return Component{ID:id,Nucleus:nucleus,Kind:kind,Capabilities:[]string{"test"},Execute:func(ctx context.Context,in []float64)([]float64,error){select{case <-ctx.Done():return nil,ctx.Err();default:};out:=append([]float64(nil),in...);for i:=range out{out[i]+=add};return out,nil}} }
	for _, c := range []Component{mk("n01.tool","N01",1,KindTool),mk("n02.agent","N02",2,KindAgent),mk("n03.tool","N03",3,KindTool)} { if err:=r.Register(c);err!=nil{t.Fatal(err)} }
	got,err:=r.Fuse(context.Background(),[]string{"n01.tool","n02.agent","n03.tool"},[]float64{0,1});if err!=nil{t.Fatal(err)}
	if got.Output[0]!=6||got.Output[1]!=7{t.Fatalf("unexpected fused output: %#v",got.Output)}
	if len(got.Trace)!=3{t.Fatalf("expected 3 trace steps, got %d",len(got.Trace))}
}
func TestFusionRejectsNonAdjacentNuclei(t *testing.T){r:=NewRegistry();mk:=func(id,n string)Component{return Component{ID:id,Nucleus:n,Kind:KindTool,Execute:func(context.Context,[]float64)([]float64,error){return []float64{1},nil}}};if err:=r.Register(mk("n01","N01"));err!=nil{t.Fatal(err)};if err:=r.Register(mk("n03","N03"));err!=nil{t.Fatal(err)};if _,err:=r.Fuse(context.Background(),[]string{"n01","n03"},[]float64{0});err==nil{t.Fatal("expected non-adjacent fusion rejection")}}
func TestFusionRejectsDuplicateRegistration(t *testing.T){r:=NewRegistry();c:=Component{ID:"x",Nucleus:"N01",Kind:KindTool,Execute:func(context.Context,[]float64)([]float64,error){return []float64{1},nil}};if err:=r.Register(c);err!=nil{t.Fatal(err)};if err:=r.Register(c);err==nil{t.Fatal("expected duplicate registration rejection")}}
