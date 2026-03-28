package inventory

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/arda/distributed-commerce-core/internal/contracts"
    "github.com/arda/distributed-commerce-core/internal/platform/inbox"
    "github.com/arda/distributed-commerce-core/internal/platform/natsx"
)

type Service struct {
    DB     *pgxpool.Pool
    Inbox  *inbox.Store
    Bus    *natsx.Client
    Logger *slog.Logger
}

func NewService(db *pgxpool.Pool, inbox *inbox.Store, bus *natsx.Client, logger *slog.Logger) *Service {
    return &Service{DB: db, Inbox: inbox, Bus: bus, Logger: logger}
}

func (s *Service) HandleReserve(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "inventory.reserve", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var cmd contracts.InventoryReserve
    if err := json.Unmarshal(env.Payload, &cmd); err != nil {
        return err
    }

    // Simple deterministic demo rule.
    if len(cmd.Items) > 0 && cmd.Items[0].Quantity > 999 {
        payload, _ := json.Marshal(contracts.InventoryRejected{OrderID: cmd.OrderID, Reason: "insufficient_stock"})
        return s.Bus.Publish(ctx, contracts.SubjectInventoryRejected, contracts.Envelope{
            EventID:       uuid.NewString(),
            EventType:     contracts.SubjectInventoryRejected,
            EventVersion:  1,
            AggregateType: "inventory",
            AggregateID:   cmd.OrderID,
            CorrelationID: env.CorrelationID,
            CausationID:   env.EventID,
            OccurredAt:    time.Now().UTC(),
            Payload:       payload,
        })
    }

    _, err = s.DB.Exec(ctx, `INSERT INTO inventory_reservations (id, order_id, status, created_at, updated_at) VALUES ($1, $2, 'RESERVED', NOW(), NOW())`, uuid.NewString(), cmd.OrderID)
    if err != nil {
        return err
    }

    payload, _ := json.Marshal(contracts.InventoryReserved{OrderID: cmd.OrderID})
    return s.Bus.Publish(ctx, contracts.SubjectInventoryReserved, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectInventoryReserved,
        EventVersion:  1,
        AggregateType: "inventory",
        AggregateID:   cmd.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    })
}
