# ADR-001: Bound retries for temporary gateway failures

- Status: Accepted
- Jira: <actual Jira key 2>
- Services: payments, gateway
- Code: internal/payments/retry.go; internal/http/charge.go
- Slack: #incident

## Context

The simulated gateway can return temporary 503 responses. Retrying forever
could send too many requests while it is unhealthy.

## Decision

The payments service retries network errors and 5xx responses at most three
times, using bounded backoff. It does not retry 4xx responses. The gateway
returns Retry-After with its temporary failure response.

## Evidence

See the simulated test and decision discussion in Slack #incident, and the
two pull requests linked to the Jira issue above.