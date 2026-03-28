package notification

import (
    "context"
    "encoding/json"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/arda/distributed-commerce-core/internal/contracts"
    "github.com/arda/distributed-commerce-core/internal/platform/inbox"
)

type Service struct {
    DB     *pgxpool.Pool
    Inbox  *inbox.Store
    Logger *slog.Logger
}

func NewService(db *pgxpool.Pool, inbox *inbox.Store, logger *slog.Logger) *Service {
    return &Service{DB: db, Inbox: inbox, Logger: logger}
}

func (s *Service) HandleSend(ctx context.Context, env contracts.Envelope) error {
    fresh, err := s.Inbox.MarkProcessed(ctx, "notification.send", env.EventID)
    if err != nil || !fresh {
        return err
    }
    var msg contracts.NotificationSend
    if err := json.Unmarshal(env.Payload, &msg); err != nil {
        return err
    }
    _, err = s.DB.Exec(ctx, `INSERT INTO notifications (id, order_id, kind, message, created_at) VALUES ($1, $2, $3, $4, NOW())`, env.EventID, msg.OrderID, msg.Kind, msg.Message)
    if err != nil {
        return err
    }
    s.Logger.Info("notification_recorded", "order_id", msg.OrderID, "kind", msg.Kind)
    return nil
}
