package supergpu

import("context";"testing")
func TestRuntimeInitializesComputeDiscovery(t *testing.T){r:=New(nil);d:=r.Discover();if len(d)==0{t.Fatal("no compute device")};if r.Health()["status"]!="ready"{t.Fatalf("runtime status=%v",r.Health()["status"])};if _,err:=r.Select("");err!=nil{t.Fatal(err)}}
func TestRuntime(t *testing.T){r:=New(nil);d:=r.Discover();if len(d)==0{t.Fatal("no compute device")};if err:=r.Reserve(d[0].ID,"test");err!=nil{t.Fatal(err)};v,err:=r.Execute(context.Background(),d[0],"square",[]float64{2,3});if err!=nil{t.Fatal(err)};if v[0]!=4||v[1]!=9{t.Fatal("compute result incorrect")};if _,err=r.Batch(context.Background(),d[0],"identity",[][]float64{{1},{2}});err!=nil{t.Fatal(err)};if err=r.Release(d[0].ID,"test");err!=nil{t.Fatal(err)};if r.Health()["status"]!="ready"{t.Fatal("runtime unhealthy")};if err=r.Shutdown();err!=nil{t.Fatal(err)}}
