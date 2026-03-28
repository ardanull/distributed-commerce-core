package outbox

import (
    "context"
    "encoding/json"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/arda/distributed-commerce-core/internal/contracts"
)

type Message struct {
    ID        string
    Subject   string
    Payload   []byte
    CreatedAt time.Time
}

type Store struct {
    DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
    return &Store{DB: db}
}

func (s *Store) Enqueue(ctx context.Context, tx pgxTx, subject string, env contracts.Envelope) error {
    payload, err := json.Marshal(env)
    if err != nil {
        return err
    }
    _, err = tx.Exec(ctx, `
        INSERT INTO outbox_messages (id, subject, payload, created_at)
        VALUES ($1, $2, $3, NOW())
    `, env.EventID, subject, payload)
    return err
}

func (s *Store) FetchBatch(ctx context.Context, limit int) ([]Message, error) {
    rows, err := s.DB.Query(ctx, `
        SELECT id, subject, payload, created_at
        FROM outbox_messages
        WHERE published_at IS NULL
        ORDER BY created_at
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Message
    for rows.Next() {
        var m Message
        if err := rows.Scan(&m.ID, &m.Subject, &m.Payload, &m.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, m)
    }
    return out, rows.Err()
}

func (s *Store) MarkPublished(ctx context.Context, id string) error {
    _, err := s.DB.Exec(ctx, `UPDATE outbox_messages SET published_at = NOW() WHERE id = $1`, id)
    return err
}

type pgxTx interface {
    Exec(context.Context, string, ...any) (any, error)
}
