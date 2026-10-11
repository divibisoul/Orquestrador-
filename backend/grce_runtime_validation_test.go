package backend

import (
	"context"
	"strings"
	"testing"

	"github.com/divibisoul/Orquestrador-/grf"
)

func TestGRCEValidateITRCountsPreservedOriginalPlusAddedEvidence(t *testing.T) {
	runtime := &GRCEExecutorRuntime{}
	original := map[string]any{"text": strings.Repeat("a", 11000)}
	candidate := grf.State{
		ID: "candidate",
		EpistemicState: grf.PROJECTED,
		Payload: map[string]any{
			"original_state": original,
			"artifacts": []grf.Artifact{{
				ID: "artifact:ara",
				Stage: "ARA",
				State: grf.REAL,
				Payload: map[string]any{"failure_id": "ara:observed", "new_evidence": strings.Repeat("e", 128)},
			}},
			"provenance": []grf.Provenance{{
				ParentHash: "parent", InputHash: "input", OutputHash: "output",
				SequenceIndex: 2, Stage: "ARA",
			}},
		},
	}

	ok, err := runtime.validateITR(context.Background(), candidate, grf.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ITR rejected a candidate that preserves the original input and adds evidence")
	}
}

func TestGRCEValidateITRRejectsMissingOriginalOrArtifacts(t *testing.T) {
	runtime := &GRCEExecutorRuntime{}
	tests := []struct {
		name    string
		payload map[string]any
		wantErr string
	}{
		{
			name: "missing original",
			payload: map[string]any{
				"artifacts": []grf.Artifact{{ID: "artifact", Stage: "ARA", State: grf.REAL}},
			},
			wantErr: "GRCE_VALIDATE_ITR_ORIGINAL_STATE_MISSING",
		},
		{
			name: "missing artifacts",
			payload: map[string]any{
				"original_state": map[string]any{"text": "preserve"},
			},
			wantErr: "GRCE_VALIDATE_ITR_ARTIFACTS_MISSING",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, err := runtime.validateITR(context.Background(), grf.State{Payload: test.payload}, grf.Context{})
			if err == nil || err.Error() != test.wantErr {
				t.Fatalf("expected %s, got ok=%t err=%v", test.wantErr, ok, err)
			}
			if ok {
				t.Fatal("malformed candidate must not pass ITR")
			}
		})
	}
}
