package backend

import (
	"context"
	"errors"
	"net/http"

	"github.com/divibisoul/Orquestrador-/rgo"
)

func (p *SARAProxy) RGOIngest(ctx context.Context, env rgo.Envelope) (map[string]any, error) {
	var out map[string]any
	if err := p.request(ctx, "POST", "/v1/rgo/ingest", env, env.CorrelationID, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (p *SARAProxy) RGOTrinity(ctx context.Context, envelope map[string]any, correlationID string) (map[string]any, error) {
	var out map[string]any
	if envelope == nil {
		return nil, errors.New("RGO envelope is required")
	}
	if err := p.request(ctx, http.MethodPost, "/v1/rgo/trinity", envelope, correlationID, &out); err != nil {
		return nil, err
	}
	return out, nil
}
