# ADR 0002: Use Transactional Outbox

## Status
Accepted

## Context
Order writes and event publication must not drift apart if a service crashes.

## Decision
Persist integration events in an outbox table inside the same transaction as aggregate changes.

## Consequences
- Event delivery becomes eventually consistent but reliable
- Requires a publisher worker
- Strongly improves failure tolerance and production realism
