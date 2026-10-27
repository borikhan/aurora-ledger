# Architecture and operations

Aurora keeps transport concerns in `cmd/api` and business invariants in `internal/ledger`. The domain owns account mutation and journal creation under one lock, so readers can never observe a half-applied transfer. Duplicate transfer IDs return the original result.

## Production evolution

Use PostgreSQL as an attached resource: lock both accounts in stable ID order, insert two journal rows, update balances, and commit in one serializable transaction. A unique transfer ID makes client retries safe. Emit an outbox record in that transaction for downstream events; a separate worker publishes it. Partition journal history by time only after measurement proves it necessary.

Key telemetry is request latency, transfer outcomes, idempotency replays, insufficient-funds decisions, and the zero-sum audit. Never place account data in logs.

## Twelve-factor mapping

| Factor | Implementation |
|---|---|
| Codebase | One Git repository, independently deployable API |
| Dependencies | Go modules and a pinned toolchain declaration |
| Config | `PORT` and future database credentials come from environment variables |
| Backing services | Persistence is behind the ledger boundary; PostgreSQL is replaceable by URL/config |
| Build, release, run | CI verifies; Docker builds an immutable binary; runtime receives config |
| Processes | API instances hold no durable production state; the memory adapter is local-only |
| Port binding | Self-contained HTTP service binds `$PORT` |
| Concurrency | Horizontal API processes; database transactions coordinate writes |
| Disposability | Fast startup, bounded HTTP headers, retry-safe transfer commands |
| Dev/prod parity | The same binary and container run in every environment |
| Logs | Structured records go to stdout/stderr |
| Admin processes | Audit or migration commands run as one-off tasks from the same release |

## Explicit trade-offs

This repository proves domain behavior, not regulatory compliance. Authentication, authorization, KMS-backed secrets, PostgreSQL migrations, reconciliation jobs, and jurisdiction-specific controls belong in a real deployment.
