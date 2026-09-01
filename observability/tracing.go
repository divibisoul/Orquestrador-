package observability

import("context";"go.opentelemetry.io/otel";"go.opentelemetry.io/otel/attribute";"go.opentelemetry.io/otel/codes";"go.opentelemetry.io/otel/trace")

var tracer=otel.Tracer("soul.n07")
func Start(ctx context.Context,operation,traceID string,payloadSize int)(context.Context,trace.Span){if ctx==nil{ctx=context.Background()};ctx,span:=tracer.Start(ctx,operation);span.SetAttributes(attribute.String("soul.nucleus","N07"),attribute.String("operation",operation),attribute.String("soul.trace_id",traceID),attribute.Int("payload.size",payloadSize));return ctx,span}
func End(span trace.Span,err error){if span==nil{return};if err!=nil{span.RecordError(err);span.SetStatus(codes.Error,err.Error())}else{span.SetStatus(codes.Ok,"")};span.End()}
