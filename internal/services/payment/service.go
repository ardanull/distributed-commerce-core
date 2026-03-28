package payment

import (
    "context"
    "encoding/json"
    "log/slog"
    "math/rand"
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

func (s *Service) HandleAuthorize(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "payment.authorize", env.EventID)
    if err != nil || !fresh {
        return err
    }

    var cmd contracts.PaymentAuthorize
    if err := json.Unmarshal(env.Payload, &cmd); err != nil {
        return err
    }

    // Simulated decision engine.
    if rand.Intn(10) == 0 {
        payload, _ := json.Marshal(contracts.PaymentRejected{OrderID: cmd.OrderID, Reason: "risk_engine_rejected"})
        return s.Bus.Publish(ctx, contracts.SubjectPaymentRejected, contracts.Envelope{
            EventID:       uuid.NewString(),
            EventType:     contracts.SubjectPaymentRejected,
            EventVersion:  1,
            AggregateType: "payment",
            AggregateID:   cmd.OrderID,
            CorrelationID: env.CorrelationID,
            CausationID:   env.EventID,
            OccurredAt:    time.Now().UTC(),
            Payload:       payload,
        })
    }

    paymentID := uuid.NewString()
    _, err = s.DB.Exec(ctx, `
        INSERT INTO payments (id, order_id, status, amount, currency, created_at, updated_at)
        VALUES ($1, $2, 'AUTHORIZED', $3, $4, NOW(), NOW())
    `, paymentID, cmd.OrderID, cmd.Amount, cmd.Currency)
    if err != nil {
        return err
    }

    payload, _ := json.Marshal(contracts.PaymentAuthorized{
        OrderID: cmd.OrderID, PaymentID: paymentID, AuthorizedAt: time.Now().UTC().Format(time.RFC3339),
    })
    return s.Bus.Publish(ctx, contracts.SubjectPaymentAuthorized, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectPaymentAuthorized,
        EventVersion:  1,
        AggregateType: "payment",
        AggregateID:   cmd.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    })
}

func (s *Service) HandleRefund(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "payment.refund", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var cmd contracts.PaymentRefund
    if err := json.Unmarshal(env.Payload, &cmd); err != nil {
        return err
    }
    _, err = s.DB.Exec(ctx, `UPDATE payments SET status = 'REFUNDED', updated_at = NOW() WHERE id = $1`, cmd.PaymentID)
    if err != nil {
        return err
    }

    payload, _ := json.Marshal(contracts.PaymentRefunded{OrderID: cmd.OrderID})
    return s.Bus.Publish(ctx, contracts.SubjectPaymentRefunded, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectPaymentRefunded,
        EventVersion:  1,
        AggregateType: "payment",
        AggregateID:   cmd.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    })
}
