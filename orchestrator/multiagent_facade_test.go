package orchestrator
import("context";"testing";"github.com/divibisoul/Orquestrador-/protocol")
func TestMultiAgentEvidenceFailClosed(t *testing.T){t.Setenv("SOUL_N07_CREWAI_ENABLED","");e:=providerEvidence("crewai");if e.State!="DEGRADED"||e.Code!="MULTIAGENT_ADAPTER_DISABLED"{t.Fatalf("%#v",e)}}
func TestMultiAgentFacadeRegister(t *testing.T){e,err:=NewTestEngine();if err!=nil{t.Fatal(err)};if err=RegisterMultiAgentFacadeOperations(e);err!=nil{t.Fatal(err)};if !e.HasOperation(MultiAgentDescribeOperation)||!e.HasOperation(MultiAgentExecuteOperation){t.Fatal("missing facade ops")}}
func TestCorrelationPreserved(t *testing.T){m:=protocol.Message{CorrelationID:"corr-test"};r,_:=multiAgentFail(m,"TEST");if r.CorrelationID!="corr-test"{t.Fatal("correlation lost")};_ = context.Background()}

func TestMultiAgentRequestFromMetadata(t *testing.T){
 m:=protocol.Message{Metadata:map[string]string{"provider":"crewai","goal":"test-goal","roles":"[\\"builder\\"]","maxRounds":"4"}}
 q:=multiAgentRequestFromMessage(m)
 if q.Provider!="crewai"||q.Goal!="test-goal"||len(q.Roles)!=1||q.Roles[0]!="builder"||q.MaxRounds!=4{t.Fatalf("unexpected request: %#v",q)}
}
