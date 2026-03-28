package inbox

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
    DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
    return &Store{DB: db}
}

func (s *Store) MarkProcessed(ctx context.Context, consumer, eventID string) (bool, error) {
    tag, err := s.DB.Exec(ctx, `
        INSERT INTO inbox_messages (consumer_name, event_id, processed_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT DO NOTHING
    `, consumer, eventID)
    if err != nil {
        return false, err
    }
    return tag.RowsAffected() == 1, nil
}
