package telemetry

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func Init(ctx context.Context, endpoint, service string, enabled bool) (func(context.Context) error, error) {
    if !enabled {
        return func(context.Context) error { return nil }, nil
    }
    exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exp),
        tracesdk.WithResource(resource.Empty()),
    )
    otel.SetTracerProvider(tp)
    return tp.Shutdown, nil
}
