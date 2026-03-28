package contracts

import (
    "encoding/json"
    "time"
)

type Envelope struct {
    EventID       string          `json:"event_id"`
    EventType     string          `json:"event_type"`
    EventVersion  int             `json:"event_version"`
    AggregateType string          `json:"aggregate_type"`
    AggregateID   string          `json:"aggregate_id"`
    CorrelationID string          `json:"correlation_id"`
    CausationID   string          `json:"causation_id"`
    ActorID       string          `json:"actor_id,omitempty"`
    Traceparent   string          `json:"traceparent,omitempty"`
    OccurredAt    time.Time       `json:"occurred_at"`
    Payload       json.RawMessage `json:"payload"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}
