# DLQ Runbook

## Symptoms
- Rising retry counts
- Messages repeatedly NAKed
- Backlog growth on dead-letter subject

## Triage
1. Inspect recent logs with the correlation ID.
2. Identify failing event type and consumer.
3. Verify DB and dependency health.
4. Determine if the payload is malformed or business-invalid.

## Recovery
- Fix the bug or dependency issue
- Replay messages from the dead-letter stream
- Confirm dedup/inbox prevents double-processing
