# Design decisions

Each entry: what we chose, and why.

## Money is stored as integer cents
Amounts are whole numbers of the smallest unit (cents). Never floats.
Why: floats round, and rounding loses money.

## One PostgreSQL database, one schema per service
Schemas: `auth` (gateway), `wallet` (wallet), `ledger` (ledger).
Why: each service owns its tables, and no service writes another's tables.
Splitting into separate databases later is easy because nothing crosses schemas.

## Kafka runs in KRaft mode, single broker, for local dev
KRaft means Kafka manages itself without ZooKeeper.
Why: one container instead of two. Production would run 3 brokers.

## transfers.completed has 16 partitions
A consumer group uses at most one consumer per partition.
Why: the plan is to autoscale the ledger to 15 pods with KEDA,
so the topic needs at least 15 partitions or the extra pods sit idle.

## Transactional outbox for events
The wallet writes the balance change and an outbox row in one transaction.
A relay loop publishes outbox rows to Kafka.
Why: without it, the balance could change while the Kafka publish fails,
and the ledger would never hear about it.

## Every consumer is idempotent
Why: Kafka delivers at least once, so the same event can arrive twice.

## The ledger is append-only and enforces debits = credits in the database
A deferred trigger rejects any unbalanced entry at COMMIT.
Another trigger rejects UPDATE and DELETE on journal tables.
Why: correctness is guaranteed by the database, not only by application code.

## Topic names use dots only, never underscores
Kafka metric names turn both "." and "_" into "_", so a.b and a_b would collide.
Why: sticking to one separator avoids the clash Kafka warns about.

## Services use the languages from the project blueprint
gateway and wallet in Go, ledger in Java (Spring Boot), fraud in Python (FastAPI),
notification in Node.js.
Why: the blueprint shows several languages behind one platform, so the deployment
side has to handle a different build and runtime for each.


## Both Go services share one Go module at the repo root
Why: both need the generated gRPC code. One go.mod lets them import it directly.

## gRPC contract lives in proto/, generated with buf, generated code is committed
Why: buf lints the contract and can detect breaking changes in CI.
Committing gen/ means Docker builds only need Go, not buf and the code generators.

## Every RPC has its own Request and Response message
Why: one response can gain fields later without touching any other call.

## Ops endpoints on a separate port (9090)
/healthz, /readyz and /metrics are served on OPS_PORT, not the main port.
Why: they should never be reachable through the ingress, and probes keep working
even if the main server is overloaded.

## Kafka is not a readiness check for the wallet
Why: if Kafka is down, transfers still succeed and wait in the outbox.
A Kafka outage should not take payments down.

## Graceful shutdown: 5s drain, then 20s cleanup
On SIGTERM: /readyz returns 503, wait SHUTDOWN_DELAY (5s), stop servers, close connections.
A timer force-exits after SHUTDOWN_DELAY + SHUTDOWN_TIMEOUT (25s total).
Why: the wait lets Kubernetes remove the pod from its Service before we stop taking work,
and 25s fits inside the default 30s terminationGracePeriodSeconds.