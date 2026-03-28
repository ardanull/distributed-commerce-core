package order

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/arda/distributed-commerce-core/internal/contracts"
    "github.com/arda/distributed-commerce-core/internal/platform/outbox"
)

type Repository struct {
    DB     *pgxpool.Pool
    Outbox *outbox.Store
}

func NewRepository(db *pgxpool.Pool, out *outbox.Store) *Repository {
    return &Repository{DB: db, Outbox: out}
}

func (r *Repository) Create(ctx context.Context, o Order, env contracts.Envelope) error {
    tx, err := r.DB.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    itemsJSON, _ := json.Marshal(o.Items)
    _, err = tx.Exec(ctx, `
        INSERT INTO orders (id, customer_id, currency, total_amount, status, items, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `, o.ID, o.CustomerID, o.Currency, o.TotalAmount, o.Status, itemsJSON)
    if err != nil {
        return err
    }

    if err := r.Outbox.Enqueue(ctx, tx, contracts.SubjectOrderCreated, env); err != nil {
        return err
    }

    return tx.Commit(ctx)
}

func (r *Repository) Get(ctx context.Context, id string) (Order, error) {
    row := r.DB.QueryRow(ctx, `SELECT id, customer_id, currency, total_amount, status, items, created_at, updated_at FROM orders WHERE id = $1`, id)
    var o Order
    var itemsJSON []byte
    if err := row.Scan(&o.ID, &o.CustomerID, &o.Currency, &o.TotalAmount, &o.Status, &itemsJSON, &o.CreatedAt, &o.UpdatedAt); err != nil {
        return Order{}, err
    }
    _ = json.Unmarshal(itemsJSON, &o.Items)
    return o, nil
}

func (r *Repository) Transition(ctx context.Context, id string, next Status) error {
    current, err := r.Get(ctx, id)
    if err != nil {
        return err
    }
    if !AllowedTransition(current.Status, next) {
        return fmt.Errorf("invalid transition %s -> %s", current.Status, next)
    }
    _, err = r.DB.Exec(ctx, `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`, id, next)
    return err
}

func NewOrder(customerID, currency string, items []Item) Order {
    var total int64
    for _, it := range items {
        total += int64(it.Quantity) * it.UnitPrice
    }
    return Order{
        ID:          uuid.NewString(),
        CustomerID:  customerID,
        Currency:    currency,
        TotalAmount: total,
        Status:      StatusPendingPayment,
        Items:       items,
        CreatedAt:   time.Now().UTC(),
        UpdatedAt:   time.Now().UTC(),
    }
}
