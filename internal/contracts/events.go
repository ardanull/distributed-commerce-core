package contracts

const (
    SubjectOrderCreated       = "order.created.v1"
    SubjectOrderComplete      = "order.complete.v1"
    SubjectOrderFail          = "order.fail.v1"
    SubjectPaymentAuthorize   = "payment.authorize.v1"
    SubjectPaymentAuthorized  = "payment.authorized.v1"
    SubjectPaymentRejected    = "payment.rejected.v1"
    SubjectPaymentRefund      = "payment.refund.v1"
    SubjectPaymentRefunded    = "payment.refunded.v1"
    SubjectInventoryReserve   = "inventory.reserve.v1"
    SubjectInventoryReserved  = "inventory.reserved.v1"
    SubjectInventoryRejected  = "inventory.rejected.v1"
    SubjectNotificationSend   = "notification.send.v1"
)

type OrderItem struct {
    SKU       string `json:"sku"`
    Quantity  int    `json:"quantity"`
    UnitPrice int64  `json:"unit_price"`
}

type OrderCreated struct {
    OrderID     string      `json:"order_id"`
    CustomerID  string      `json:"customer_id"`
    Currency    string      `json:"currency"`
    TotalAmount int64       `json:"total_amount"`
    Items       []OrderItem `json:"items"`
}

type PaymentAuthorize struct {
    OrderID     string `json:"order_id"`
    CustomerID  string `json:"customer_id"`
    Amount      int64  `json:"amount"`
    Currency    string `json:"currency"`
}

type PaymentAuthorized struct {
    OrderID      string `json:"order_id"`
    PaymentID    string `json:"payment_id"`
    AuthorizedAt string `json:"authorized_at"`
}

type PaymentRejected struct {
    OrderID string `json:"order_id"`
    Reason  string `json:"reason"`
}

type PaymentRefund struct {
    OrderID   string `json:"order_id"`
    PaymentID string `json:"payment_id"`
    Reason    string `json:"reason"`
}

type PaymentRefunded struct {
    OrderID string `json:"order_id"`
}

type InventoryReserve struct {
    OrderID string      `json:"order_id"`
    Items   []OrderItem `json:"items"`
}

type InventoryReserved struct {
    OrderID string `json:"order_id"`
}

type InventoryRejected struct {
    OrderID string `json:"order_id"`
    Reason  string `json:"reason"`
}

type OrderComplete struct {
    OrderID string `json:"order_id"`
}

type OrderFail struct {
    OrderID string `json:"order_id"`
    Reason  string `json:"reason"`
}

type NotificationSend struct {
    OrderID string `json:"order_id"`
    Kind    string `json:"kind"`
    Message string `json:"message"`
}
