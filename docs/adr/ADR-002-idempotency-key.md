# ADR-002: Require an idempotency key for payment creation

- Status: Accepted
- Jira: <actual Jira key 3>
- Service: payments
- Code: internal/payments/idempotency.go
- Slack: #decisions

## Context

A client may retry after a timeout without knowing whether its first request
succeeded. Treating the retry as new could create a duplicate payment.

## Decision

Require an Idempotency-Key on POST /payments. A repeated request with the
same key returns the original fake payment ID.

## Consequences

The demo can safely handle repeated requests. Its in-memory records reset
when the service restarts.

## Evidence

See the discussion in Slack #decisions and the pull request linked to the
Jira issue above.