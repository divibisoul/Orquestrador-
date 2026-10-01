package supergpu

import "context"

type correlationContextKey struct{}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, correlationContextKey{}, correlationID)
}

func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(correlationContextKey{}).(string)
	return value
}

type ExecutionEvent struct {
	Phase         string
	Operation     string
	DeviceID      string
	Backend       string
	InputSize     int
	OutputSize    int
	CorrelationID string
	Error         string
}

type ExecutionReporter interface {
	Report(context.Context, ExecutionEvent) error
}

type ReporterFunc func(context.Context, ExecutionEvent) error

func (f ReporterFunc) Report(ctx context.Context, event ExecutionEvent) error {
	return f(ctx, event)
}
