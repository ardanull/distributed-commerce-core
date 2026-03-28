# Workflow

```text
order.created -> payment.authorize -> payment.authorized -> inventory.reserve -> inventory.reserved -> order.complete -> notification.send
                                                                \-> inventory.rejected -> payment.refund -> payment.refunded -> order.fail -> notification.send
```
