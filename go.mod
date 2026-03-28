module github.com/arda/distributed-commerce-core

go 1.23.2

require (
    github.com/go-chi/chi/v5 v5.2.1
    github.com/go-redis/redis/v8 v8.11.5
    github.com/google/uuid v1.6.0
    github.com/jackc/pgx/v5 v5.7.4
    github.com/nats-io/nats.go v1.47.0
    github.com/prometheus/client_golang v1.23.2
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0
    go.opentelemetry.io/otel v1.38.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
    go.opentelemetry.io/otel/sdk v1.38.0
    golang.org/x/sync v0.17.0
)
