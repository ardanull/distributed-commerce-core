package orchestrator

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/google/uuid"

    "github.com/arda/distributed-commerce-core/internal/contracts"
    "github.com/arda/distributed-commerce-core/internal/platform/inbox"
    "github.com/arda/distributed-commerce-core/internal/platform/natsx"
)

type Service struct {
    Inbox  *inbox.Store
    Bus    *natsx.Client
    Logger *slog.Logger
}

func NewService(inbox *inbox.Store, bus *natsx.Client, logger *slog.Logger) *Service {
    return &Service{Inbox: inbox, Bus: bus, Logger: logger}
}

func (s *Service) OnOrderCreated(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.order_created", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.OrderCreated
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }

    payload, _ := json.Marshal(contracts.PaymentAuthorize{
        OrderID: event.OrderID, CustomerID: event.CustomerID, Amount: event.TotalAmount, Currency: event.Currency,
    })
    return s.Bus.Publish(ctx, contracts.SubjectPaymentAuthorize, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectPaymentAuthorize,
        EventVersion:  1,
        AggregateType: "order",
        AggregateID:   event.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    })
}

func (s *Service) OnPaymentAuthorized(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.payment_authorized", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.PaymentAuthorized
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }
    payload, _ := json.Marshal(contracts.InventoryReserve{OrderID: event.OrderID})
    return s.Bus.Publish(ctx, contracts.SubjectInventoryReserve, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectInventoryReserve,
        EventVersion:  1,
        AggregateType: "order",
        AggregateID:   event.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    })
}

func (s *Service) OnPaymentRejected(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.payment_rejected", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.PaymentRejected
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }
    return s.failOrder(ctx, env, event.OrderID, event.Reason)
}

func (s *Service) OnInventoryReserved(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.inventory_reserved", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.InventoryReserved
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }
    payload, _ := json.Marshal(contracts.OrderComplete{OrderID: event.OrderID})
    if err := s.Bus.Publish(ctx, contracts.SubjectOrderComplete, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectOrderComplete,
        EventVersion:  1,
        AggregateType: "order",
        AggregateID:   event.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    }); err != nil {
        return err
    }
    notify, _ := json.Marshal(contracts.NotificationSend{
        OrderID: event.OrderID, Kind: "success", Message: "Order completed successfully",
    })
    return s.Bus.Publish(ctx, contracts.SubjectNotificationSend, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectNotificationSend,
        EventVersion:  1,
        AggregateType: "notification",
        AggregateID:   event.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       notify,
    })
}

func (s *Service) OnInventoryRejected(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.inventory_rejected", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.InventoryRejected
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }

    // In a real implementation we would lookup payment_id. Here we use placeholder "latest".
    refund, _ := json.Marshal(contracts.PaymentRefund{
        OrderID: event.OrderID, PaymentID: "latest", Reason: event.Reason,
    })
    return s.Bus.Publish(ctx, contracts.SubjectPaymentRefund, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectPaymentRefund,
        EventVersion:  1,
        AggregateType: "payment",
        AggregateID:   event.OrderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       refund,
    })
}

func (s *Service) OnPaymentRefunded(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "orchestrator.payment_refunded", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var event contracts.PaymentRefunded
    if err := json.Unmarshal(env.Payload, &event); err != nil {
        return err
    }
    return s.failOrder(ctx, env, event.OrderID, "payment refunded after inventory rejection")
}

func (s *Service) failOrder(ctx context.Context, env contracts.Envelope, orderID, reason string) error {
    payload, _ := json.Marshal(contracts.OrderFail{OrderID: orderID, Reason: reason})
    if err := s.Bus.Publish(ctx, contracts.SubjectOrderFail, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectOrderFail,
        EventVersion:  1,
        AggregateType: "order",
        AggregateID:   orderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       payload,
    }); err != nil {
        return err
    }
    notify, _ := json.Marshal(contracts.NotificationSend{
        OrderID: orderID, Kind: "failure", Message: reason,
    })
    return s.Bus.Publish(ctx, contracts.SubjectNotificationSend, contracts.Envelope{
        EventID:       uuid.NewString(),
        EventType:     contracts.SubjectNotificationSend,
        EventVersion:  1,
        AggregateType: "notification",
        AggregateID:   orderID,
        CorrelationID: env.CorrelationID,
        CausationID:   env.EventID,
        OccurredAt:    time.Now().UTC(),
        Payload:       notify,
    })
}
