package order

import "testing"

func TestAllowedTransition(t *testing.T) {
    if !AllowedTransition(StatusPendingPayment, StatusPaymentAuthorized) {
        t.Fatal("expected transition to be allowed")
    }
    if AllowedTransition(StatusCompleted, StatusPendingPayment) {
        t.Fatal("expected transition to be rejected")
    }
}
