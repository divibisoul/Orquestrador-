package mesh

import (
	"strings"
	"testing"
	"time"
)

func TestCapabilityInDescriptionRequiresExecutableCapabilities(t *testing.T) {
	declarativeOnly := map[string]any{"operations": []any{"audio.transcribe@1.0.0", "tool.run@1.2.0"}}
	if capabilityInDescription(declarativeOnly, "audio.transcribe") {
		t.Fatal("declarative operations must not prove executability")
	}

	executable := map[string]any{"executableCapabilities": []any{"audio.transcribe@1.0.0", "tool.run@1.2.0"}}
	if !capabilityInDescription(executable, "audio.transcribe") {
		t.Fatal("expected executable capability to be discoverable")
	}
	if !capabilityInDescription(executable, "audio.transcribe@1.0.0") {
		t.Fatal("expected exact executable version to be discoverable")
	}
	if capabilityInDescription(executable, "audio.transcribe@2.0.0") {
		t.Fatal("unexpected incompatible executable version")
	}
	if capabilityInDescription(executable, "audio.synthesize") {
		t.Fatal("unexpected undiscovered capability")
	}
}

func TestIOReadLimitedRejectsOversizePayload(t *testing.T) {
	if _, err := ioReadLimited(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("expected oversized payload to be rejected")
	}
}

func TestFederatedParallelConstants(t *testing.T) {
	if maxFederatedParallelTasks < 1 {
		t.Fatal("parallel limit must be positive")
	}
	if federatedTaskTimeout <= 0 {
		t.Fatal("task timeout must be positive")
	}
	if federatedTaskTimeout > time.Hour {
		t.Fatal("task timeout unreasonably large")
	}
}
