# Distributed Commerce Core

A production-grade, event-driven commerce backend written in Go. This repository is designed as a **staff/principal-caliber portfolio project** and showcases:

- Event-driven microservices in Go
- Transactional **Outbox** and consumer **Inbox/Dedup**
- **Saga Orchestrator** for long-running workflows
- **NATS JetStream** for durable messaging
- **PostgreSQL** for service state and outbox storage
- **Redis** for idempotency and hot read caching
- **OpenTelemetry**, **Prometheus**, **Grafana**, **Tempo**
- **Correlation IDs**, **Causation IDs**, request tracing, structured logs
- **Dead-letter queues**, retries with backoff, poison message handling
- **Kubernetes** manifests and Docker Compose
- **k6** load testing, runbooks, ADRs, and CI

## Architecture

```text
                        +--------------------+
                        |     API Gateway    |
                        | auth / limits /    |
                        | trace propagation  |
                        +---------+----------+
                                  |
                                  v
                        +---------+----------+
                        |    Order Service   |
                        | order aggregate +  |
                        | outbox publisher   |
                        +---------+----------+
                                  |
                                  | order.created
                                  v
                    +-------------+---------------+
                    |         Saga Orchestrator   |
                    | command choreography guard  |
                    +------+------+------+--------+
                           |      |      |
                 pay.authorize    |   inventory.reserve
                           |      |      |
                           v      |      v
                    +------+--+   |   +--+----------------+
                    | Payment |   |   | Inventory Service |
                    | Service |   |   | reservations      |
                    +----+----+   |   +--------+----------+
                         |        |            |
                         | payment.*           | inventory.*
                         +--------+------------+
                                  |
                                  v
                        +---------+----------+
                        | Notification Svc   |
                        | email/webhook/audit|
                        +--------------------+
```

## Core design choices

1. **Outbox Pattern**: aggregate writes and integration events are committed atomically.
2. **Inbox / Dedup Pattern**: every consumer tracks processed event IDs and behaves effectively-once.
3. **Saga Orchestration**: order lifecycle is explicit and observable, including compensation.
4. **Observability First**: every HTTP request, DB call, message publish, and message consume can be traced.
5. **Operational Maturity**: dashboards, runbooks, SLOs, and load tests are part of the repo.

## Services

- **gateway**: public entry point, request validation, correlation/tracing, JWT placeholder, rate limiting.
- **order-service**: owns the order aggregate, order state machine, outbox persistence.
- **payment-service**: authorizes/captures/refunds payments, inbox dedupe, retry-aware publishing.
- **inventory-service**: stock reservations and releases.
- **notification-service**: terminal workflow notifications and audit events.
- **orchestrator-service**: saga workflow driver and compensation coordinator.

## Workflow

### Happy path

1. `POST /v1/orders`
2. Order Service creates order as `PENDING_PAYMENT`
3. Order Service stores `order.created.v1` in outbox
4. Publisher relays outbox event to JetStream
5. Orchestrator consumes `order.created.v1`
6. Orchestrator sends `payment.authorize.v1`
7. Payment succeeds and emits `payment.authorized.v1`
8. Orchestrator sends `inventory.reserve.v1`
9. Inventory succeeds and emits `inventory.reserved.v1`
10. Orchestrator sends `order.complete.v1`
11. Order becomes `COMPLETED`
12. Notification Service emits terminal notification / audit trail

### Compensation path

If inventory reservation fails after payment authorization:

1. Orchestrator receives `inventory.rejected.v1`
2. Orchestrator emits `payment.refund.v1`
3. Payment emits `payment.refunded.v1`
4. Orchestrator emits `order.fail.v1`
5. Notification Service records the failure

## Repository layout

```text
api/
  openapi/
cmd/
  gateway/
  order-service/
  payment-service/
  inventory-service/
  notification-service/
  orchestrator-service/
internal/
  contracts/
  platform/
  services/
deploy/
  docker/
  k8s/
docs/
  adr/
  dashboards/
  diagrams/
  runbooks/
  slo/
migrations/
scripts/
test/
```

## Local quick start

```bash
make bootstrap
make up
make migrate
make smoke
```

Then create an order:

```bash
curl -X POST http://localhost:8080/v1/orders \
  -H 'Content-Type: application/json' \
  -H 'X-Correlation-ID: demo-corr-123' \
  -d '{
    "customer_id":"cust-1",
    "currency":"TRY",
    "items":[
      {"sku":"keyboard","quantity":1,"unit_price":249900},
      {"sku":"mouse","quantity":1,"unit_price":149900}
    ]
  }'
```

## Ports

- gateway: `8080`
- order-service: `8081`
- payment-service: `8082`
- inventory-service: `8083`
- notification-service: `8084`
- orchestrator-service: `8085`
- Prometheus: `9090`
- Grafana: `3000`
- Tempo: `3200`
- NATS: `4222`
- PostgreSQL: `5432`
- Redis: `6379`

## What makes this ultra-grade

- Transactional outbox and event relay workers
- Inbox tables and dedup with optimistic semantics
- Explicit order state machine enforcement
- Saga orchestration with compensation
- Typed event envelopes with versioned event names
- Trace context, correlation IDs, causation IDs, actor metadata
- Kubernetes manifests with health/readiness
- Dashboard JSON, runbooks, ADRs, SLOs, and k6 tests
- Testcontainer-oriented test scaffolding and CI

## Notes

This repository was generated in an offline environment, so external dependencies could not be downloaded and a full build could not be executed here. The code, structure, and configs are laid out to be professional and consistent, but you should run:

```bash
go mod tidy
make bootstrap
make test
docker compose up --build
```

on your machine to fetch dependencies and validate runtime behavior.
