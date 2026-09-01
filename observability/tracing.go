package observability

import "context"

type Span struct {
	operation string
	traceID   string
}

func Start(ctx context.Context, operation, traceID string, _ int) (context.Context, *Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, &Span{operation: operation, traceID: traceID}
}

func End(_ *Span, _ error) {}
