package order

import "time"

type Status string

const (
    StatusPendingPayment    Status = "PENDING_PAYMENT"
    StatusPaymentAuthorized Status = "PAYMENT_AUTHORIZED"
    StatusInventoryReserved Status = "INVENTORY_RESERVED"
    StatusCompleted         Status = "COMPLETED"
    StatusFailed            Status = "FAILED"
)

type Item struct {
    SKU       string `json:"sku"`
    Quantity  int    `json:"quantity"`
    UnitPrice int64  `json:"unit_price"`
}

type Order struct {
    ID          string    `json:"id"`
    CustomerID  string    `json:"customer_id"`
    Currency    string    `json:"currency"`
    TotalAmount int64     `json:"total_amount"`
    Status      Status    `json:"status"`
    Items       []Item    `json:"items"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func AllowedTransition(from, to Status) bool {
    allowed := map[Status]map[Status]bool{
        StatusPendingPayment: {
            StatusPaymentAuthorized: true,
            StatusFailed:            true,
        },
        StatusPaymentAuthorized: {
            StatusInventoryReserved: true,
            StatusFailed:            true,
        },
        StatusInventoryReserved: {
            StatusCompleted: true,
        },
    }
    return allowed[from][to]
}
