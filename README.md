# wallet-ledger

A digital wallet and double-entry ledger built as microservices.
Users hold balances, send money to each other, and get live notifications.
Every money movement is recorded in an append-only double-entry ledger,
and a fraud worker screens transfers in real time.

## Services

| Service      | Language | Job                                              |
|--------------|----------|--------------------------------------------------|
| gateway      | Go       | Auth (JWT), rate limiting, REST API, calls wallet over gRPC |
| wallet       | Go       | Balances, transfers, optimistic locking, outbox  |
| ledger       | TBD      | Consumes events, writes double-entry journal     |
| fraud        | Python   | Consumes transfers, amount and velocity rules    |
| notification | Node.js  | WebSocket push and webhook alerts                |

## Architecture

```
       [ Client ]
           │
     [ gateway ] ──gRPC──► [ wallet ] ──► PostgreSQL
                               │
                         (outbox relay)
                               ▼
                            Kafka
               ┌───────────────┼───────────────┐
               ▼               ▼               ▼
          [ ledger ]       [ fraud ]     [ notification ]
          PostgreSQL        Redis          WebSocket
```

## Status

In progress. See `docs/` for design decisions and problems solved along the way.