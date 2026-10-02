package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/divibisoul/Orquestrador-/agentarsenal"
	"github.com/divibisoul/Orquestrador-/protocol"
)

func boundedAgentArsenalInt(raw string, fallback, min, max int) int {
	value := fallback
	if parsed, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		value = parsed
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func agentArsenalError(m protocol.Message, err error) (protocol.Result, error) {
	return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "error", Error: err.Error()}, err
}

func RegisterAgentArsenalOperations(e *Engine, proxy *agentarsenal.Proxy) error {
	if e == nil {
		return errors.New("orchestrator engine is required")
	}
	if proxy == nil {
		return errors.New("agent arsenal proxy is required")
	}

	if err := e.Register("agent.arsenal.inventory@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		result, err := proxy.Catalog(ctx, "", "", 1000, 0)
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"inventory_json": string(raw)}}, nil
	}); err != nil {
		return err
	}

	if err := e.Register("agent.arsenal.catalog@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		source := strings.TrimSpace(m.Metadata["source"])
		kind := strings.TrimSpace(m.Metadata["kind"])
		limit := boundedAgentArsenalInt(m.Metadata["limit"], 100, 1, 1000)
		offset := boundedAgentArsenalInt(m.Metadata["offset"], 0, 0, 1000000)
		result, err := proxy.Catalog(ctx, source, kind, limit, offset)
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"catalog_json": string(raw), "source": source, "kind": kind}}, nil
	}); err != nil {
		return err
	}

	if err := e.Register("agent.arsenal.resolve@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		source := strings.TrimSpace(m.Metadata["source"])
		artifactPath := strings.TrimSpace(m.Metadata["path"])
		if source == "" || artifactPath == "" {
			return agentArsenalError(m, errors.New("metadata.source and metadata.path are required"))
		}
		result, err := proxy.Resolve(ctx, source, artifactPath)
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"artifact_json": string(raw), "source": source, "path": artifactPath}}, nil
	}); err != nil {
		return err
	}

	if err := e.Register("agent.arsenal.activate@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		source := strings.TrimSpace(m.Metadata["source"])
		artifactPath := strings.TrimSpace(m.Metadata["path"])
		task := strings.TrimSpace(m.Metadata["task"])
		if source == "" || artifactPath == "" || task == "" {
			return agentArsenalError(m, errors.New("metadata.source, metadata.path and metadata.task are required"))
		}
		result, err := proxy.Activate(ctx, source, artifactPath, task, m.Metadata["strategy"], m.Metadata["priority"])
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"activation_json": string(raw)}}, nil
	}); err != nil {
		return err
	}

	if err := e.Register("agent.arsenal.swarm@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		task := strings.TrimSpace(m.Metadata["task"])
		if task == "" {
			return agentArsenalError(m, errors.New("metadata.task is required"))
		}
		result, err := proxy.Swarm(ctx, task, m.Metadata["strategy"], m.Metadata["priority"])
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"swarm_json": string(raw)}}, nil
	}); err != nil {
		return err
	}

	return e.Register("agent.arsenal.agent.spawn@1.0.0", func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
		agentType := strings.TrimSpace(m.Metadata["type"])
		if agentType == "" {
			return agentArsenalError(m, errors.New("metadata.type is required"))
		}
		result, err := proxy.Spawn(ctx, agentType, m.Metadata["name"])
		if err != nil {
			return agentArsenalError(m, err)
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return agentArsenalError(m, err)
		}
		return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07.agent-arsenal", Target: m.Source, Status: "ok", Metadata: map[string]string{"agent_json": string(raw), "agent_type": agentType}}, nil
	})
}
