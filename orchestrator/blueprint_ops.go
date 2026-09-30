package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"strconv"

	"github.com/divibisoul/Orquestrador-/blueprint"
	"github.com/divibisoul/Orquestrador-/octacore"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func RegisterBlueprintOperations(e *Engine, processor *octacore.Processor) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if processor == nil {
		return errors.New("octacore processor is required")
	}
	if err := blueprint.ValidateManifest(); err != nil {
		return err
	}
	if err := e.Register("blueprint.resolve@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		query := strings.TrimSpace(m.Metadata["query"])
		limit := 8
		if raw := strings.TrimSpace(m.Metadata["limit"]); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				return protocol.Result{}, errors.New("metadata.limit must be a positive integer")
			}
			limit = parsed
		}
		matches, err := processor.ResolveBlueprint(query, limit)
		raw, _ := json.Marshal(matches)
		result := protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.blueprint", Target: m.Source, Status: "ok", Metadata: map[string]string{"matches_json": string(raw), "count": strconv.Itoa(len(matches)), "evidence_rule": "source-code presence never implies runtime verification"}}
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
		}
		return result, err
	}); err != nil {
		return err
	}
	if err := e.Register("blueprint.resolve.executable@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		query := strings.TrimSpace(m.Metadata["query"])
		limit := 8
		if raw := strings.TrimSpace(m.Metadata["limit"]); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				return protocol.Result{}, errors.New("metadata.limit must be a positive integer")
			}
			limit = parsed
		}
		matches, err := processor.ResolveExecutableBlueprint(query, limit)
		raw, _ := json.Marshal(matches)
		result := protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.blueprint", Target: m.Source, Status: "ok", Metadata: map[string]string{"matches_json": string(raw), "count": strconv.Itoa(len(matches)), "execution_filter": "implemented|local_executable|executable|builtin_runtime|real"}}
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
		}
		return result, err
	}); err != nil {
		return err
	}
	return e.Register("blueprint.compose@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		if err := ctx.Err(); err != nil {
			return protocol.Result{}, err
		}
		query := strings.TrimSpace(m.Metadata["query"])
		limit := 8
		if raw := strings.TrimSpace(m.Metadata["limit"]); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 {
				return protocol.Result{}, errors.New("metadata.limit must be a positive integer")
			}
			limit = parsed
		}
		plan, err := processor.ComposeBlueprint(query, limit)
		raw, _ := json.Marshal(plan)
		result := protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.blueprint", Target: m.Source, Status: "ok", Metadata: map[string]string{"route_plan_json": string(raw), "preservation": "additive; existing owners and capabilities are retained"}}
		if err != nil {
			result.Status = "error"
			result.Error = err.Error()
		}
		return result, err
	})
}
