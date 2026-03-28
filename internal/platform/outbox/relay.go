package outbox

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/arda/distributed-commerce-core/internal/contracts"
    "github.com/arda/distributed-commerce-core/internal/platform/metrics"
    "github.com/arda/distributed-commerce-core/internal/platform/natsx"
)

type Relay struct {
    Store   *Store
    Bus     *natsx.Client
    Logger  *slog.Logger
    Service string
}

func (r *Relay) Run(ctx context.Context) error {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            batch, err := r.Store.FetchBatch(ctx, 100)
            if err != nil {
                return err
            }
            for _, msg := range batch {
                var env contracts.Envelope
                if err := json.Unmarshal(msg.Payload, &env); err != nil {
                    r.Logger.Error("outbox_unmarshal_failed", "id", msg.ID, "error", err)
                    continue
                }
                if err := r.Bus.Publish(ctx, msg.Subject, env); err != nil {
                    r.Logger.Error("outbox_publish_failed", "id", msg.ID, "subject", msg.Subject, "error", err)
                    continue
                }
                metrics.OutboxPublishes.WithLabelValues(r.Service, msg.Subject).Inc()
                if err := r.Store.MarkPublished(ctx, msg.ID); err != nil {
                    return err
                }
            }
        }
    }
}
