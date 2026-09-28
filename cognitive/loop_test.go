package cognitive

import (
 "context"
 "testing"
 "time"

 "github.com/divibisoul/Orquestrador-/prefrontal"
)

type testExecutor struct{ calls int }
func(e *testExecutor) Execute(_ context.Context,capability string,payload map[string]any,correlation string)(map[string]any,string,error){
 e.calls++
 return map[string]any{"capability":capability,"payload":payload,"correlation_id":correlation},"N04",nil
}

func TestWorkingMemoryTTLAndBound(t *testing.T){
 cfg:=DefaultConfig();cfg.WorkingMemoryItems=2;cfg.WorkingMemoryTTL=10*time.Millisecond
 m:=NewWorkingMemory(cfg)
 if err:=m.Put("a",map[string]any{"v":1},.1);err!=nil{t.Fatal(err)}
 if err:=m.Put("b",map[string]any{"v":2},.2);err!=nil{t.Fatal(err)}
 if err:=m.Put("c",map[string]any{"v":3},.3);err!=nil{t.Fatal(err)}
 if len(m.Snapshot())!=2{t.Fatalf("expected bounded memory")}
 time.Sleep(15*time.Millisecond)
 if len(m.Snapshot())!=0{t.Fatalf("expected TTL eviction")}
}

func TestCritiqueBlocksHighRisk(t *testing.T){
 cortex,err:=prefrontal.New(.1,8);if err!=nil{t.Fatal(err)}
 critic,err:=NewCritic(cortex,nil,DefaultConfig());if err!=nil{t.Fatal(err)}
 err=critic.Check(context.Background(),Goal{ID:"g",Objective:"write",Risk:1,Cost:0,CorrelationID:"c"})
 if err==nil{t.Fatal("expected policy block")}
}

func TestRunPlannedFailsClosedWithoutDiscovery(t *testing.T){
 cortex,_:=prefrontal.New(.1,8)
 critic,_:=NewCritic(cortex,nil,DefaultConfig())
 exec:=&testExecutor{}
 loop,err:=New(Config{Enabled:true,GoalTTL:time.Minute,WorkingMemoryTTL:time.Minute,WorkingMemoryItems:4},NewGoalStore(),NewWorkingMemory(DefaultConfig()),exec,critic,nil)
 if err!=nil{t.Fatal(err)}
 _,err=loop.Run(context.Background(),Goal{ID:"g",Objective:"x",Capabilities:[]string{"unregistered.capability"},CorrelationID:"c",ExpiresAt:time.Now().Add(time.Minute)})
 if err==nil{t.Fatal("expected execution result or discovery guard")}
}
