package mesh

import(
	"strings"
	"testing"
	"time"
)
func TestCapabilityInDescriptionAcceptsExecutableOperations(t *testing.T){description:=map[string]any{"operations":[]any{"audio.transcribe@1.0.0","tool.run@1.2.0"}};if !capabilityInDescription(description,"audio.transcribe"){t.Fatal("expected versioned operation to be discoverable")};if capabilityInDescription(description,"audio.synthesize"){t.Fatal("unexpected undiscovered capability")}}
func TestCapabilityInDescriptionAcceptsExecutableCapabilities(t *testing.T){description:=map[string]any{"executableCapabilities":[]any{"neural.forward","compute.execute@1.0.0"}};if !capabilityInDescription(description,"neural.forward"){t.Fatal("expected executable capability to be discoverable")};if !capabilityInDescription(description,"compute.execute"){t.Fatal("expected versioned executable capability to be discoverable")}}
func TestIOReadLimitedRejectsOversizePayload(t *testing.T){if _,err:=ioReadLimited(strings.NewReader("12345"),4);err==nil{t.Fatal("expected oversized payload to be rejected")}}
func TestFederatedParallelConstants(t *testing.T){if maxFederatedParallelTasks<1{t.Fatal("parallel limit must be positive")};if federatedTaskTimeout<=0{t.Fatal("task timeout must be positive")};if federatedTaskTimeout>time.Hour{t.Fatal("task timeout unreasonably large")}}
