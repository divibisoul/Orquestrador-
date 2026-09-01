package mesh

import "testing"

func TestCapabilityInDescriptionAcceptsExecutableOperations(t *testing.T) {
	description := map[string]any{
		"operations": []any{"audio.transcribe@1.0.0", "tool.run@1.2.0"},
	}
	if !capabilityInDescription(description, "audio.transcribe") {
		t.Fatal("expected versioned operation to be discoverable")
	}
	if capabilityInDescription(description, "audio.synthesize") {
		t.Fatal("unexpected undiscovered capability")
	}
}

func TestCapabilityInDescriptionAcceptsExecutableCapabilities(t *testing.T) {
	description := map[string]any{
		"executableCapabilities": []any{"neural.forward", "compute.execute@1.0.0"},
	}
	if !capabilityInDescription(description, "neural.forward") {
		t.Fatal("expected executable capability to be discoverable")
	}
	if !capabilityInDescription(description, "compute.execute") {
		t.Fatal("expected versioned executable capability to be discoverable")
	}
}

func TestIOReadLimitedRejectsOversizePayload(t *testing.T) {
	if _, err := ioReadLimited(stringReader("12345"), 4); err == nil {
		t.Fatal("expected oversized payload to be rejected")
	}
}

type stringReader string

func (s stringReader) Read(p []byte) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}
	n := copy(p, []byte(s))
	s = s[n:]
	return n, nil
}
