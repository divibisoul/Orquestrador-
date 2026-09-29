package backend

import (
	"context"

	"github.com/divibisoul/Orquestrador-/rgo"
)

func (p *SARAProxy) RGOIngest(ctx context.Context, env rgo.Envelope) (map[string]any, error) {
	var out map[string]any
	if err := p.request(ctx, "POST", "/v1/rgo/ingest", env, env.CorrelationID, &out); err != nil {
		return nil, err
	}
	return out, nil
}
