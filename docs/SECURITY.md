# Security

## Zero Trust

Every node is treated as untrusted until authenticated and authorized. Production transport should use mTLS with workload identities such as SPIFFE/SPIRE.

## Policy boundary

`request → authentication → authorization → policy → execution → audit`.

Model output is untrusted data. Generated code and external side effects must execute inside a sandbox with least privilege.

## Secrets

Secrets must never be committed. CI should use platform secret stores. Logs must redact credentials, tokens and private payloads.

## Audit

Each task carries a trace ID and immutable event identifiers so decisions, delegations, retries and failures can be reconstructed.
