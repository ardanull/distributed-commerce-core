package natsx

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/nats-io/nats.go"

    "github.com/arda/distributed-commerce-core/internal/contracts"
)

type Client struct {
    Conn *nats.Conn
    JS   nats.JetStreamContext
}

func Connect(url string) (*Client, error) {
    nc, err := nats.Connect(url, nats.Name("distributed-commerce-core"))
    if err != nil {
        return nil, err
    }
    js, err := nc.JetStream()
    if err != nil {
        return nil, err
    }
    return &Client{Conn: nc, JS: js}, nil
}

func (c *Client) EnsureStream() error {
    _, err := c.JS.AddStream(&nats.StreamConfig{
        Name:      "COMMERCE",
        Subjects:  []string{"*.>", "order.>", "payment.>", "inventory.>", "notification.>"},
        Retention: nats.LimitsPolicy,
        Storage:   nats.FileStorage,
    })
    if err != nil && err != nats.ErrStreamNameAlreadyInUse {
        return err
    }
    return nil
}

func (c *Client) Publish(ctx context.Context, subject string, env contracts.Envelope) error {
    payload, err := json.Marshal(env)
    if err != nil {
        return err
    }
    _, err = c.JS.PublishMsg(&nats.Msg{Subject: subject, Data: payload})
    return err
}

func (c *Client) SubscribeDurable(subject, durable string, fn func(context.Context, contracts.Envelope) error) (*nats.Subscription, error) {
    return c.JS.Subscribe(subject, func(msg *nats.Msg) {
        var env contracts.Envelope
        if err := json.Unmarshal(msg.Data, &env); err != nil {
            _ = msg.Nak()
            return
        }
        if err := fn(context.Background(), env); err != nil {
            _ = msg.Nak()
            return
        }
        _ = msg.Ack()
    }, nats.Durable(durable), nats.ManualAck(), nats.AckExplicit())
}

func AckHeaderKey() string {
    return fmt.Sprintf("X-%s", "Ack-Policy")
}
