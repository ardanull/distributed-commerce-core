package order

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/google/uuid"

    "github.com/arda/distributed-commerce-core/internal/contracts"
)

type Service struct {
    Repo   *Repository
    Logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
    return &Service{Repo: repo, Logger: logger}
}

func (s *Service) Create(ctx context.Context, customerID, currency, correlationID string, items []Item) (Order, error) {
    o := NewOrder(customerID, currency, items)
    payload, _ := json.Marshal(contracts.OrderCreated{
        OrderID:     o.ID,
        CustomerID:  o.CustomerID,
        Currency:    o.Currency,
        TotalAmount: o.TotalAmount,
    })
    env := contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectOrderCreated,
        EventVersion:  1,
        AggregateType: "order",
        AggregateID:   o.ID,
        CorrelationID: correlationID,
        CausationID:   o.ID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    }
    if err := s.Repo.Create(ctx, o, env); err != nil {
        return Order{}, err
    }
    s.Logger.Info("order_created", "order_id", o.ID, "correlation_id", correlationID)
    return o, nil
}
