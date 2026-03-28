# ADR 0001: Use NATS JetStream

## Status
Accepted

## Context
The system requires durable asynchronous messaging with simple local development and low operational friction.

## Decision
Use NATS JetStream as the event backbone.

## Consequences
- Fast local startup
- Durable streams and consumer acknowledgements
- Straightforward event replay
- Not as feature-heavy as Kafka, but simpler for a portfolio platform
