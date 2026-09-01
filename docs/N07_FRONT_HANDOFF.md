# N07 front handoff

## WHAT_CHANGED

- N07 remains the final orchestration/fusion nucleus and is being evolved against the live GitHub state of N01-N06.
- Mesh gateway HMAC canonicalization was corrected after CI exposed a real `go vet` struct-field mismatch.
- Header-authenticated Mesh requests now have source+nonce replay protection with an expiry window.
- N07 Mesh preserves the canonical `correlationId` and protocol contract version while translating the wire envelope into the N07 execution message.
- N07 exposes `/health`, `/status`, `/metrics`, `/execute` and the canonical `/api/soul-mesh` gateway.
- N07 peer transport targets N01-N06 and records peer health/latency/error state.

## WHAT_WAS_FOUND

- A previous Mesh HMAC implementation used positional initialization for a struct whose field order had diverged; CI correctly caught it. The implementation now uses named fields.
- Header-HMAC authentication was cryptographically checked but did not consume the nonce through the same replay guard as envelope HMAC verification. This created a replay window and was corrected.
- The N07 protocol was aligned to Mesh contract `1.1.0` so the orchestration message carries `correlation_id` as a first-class field.

## WHAT_REMAINS

- Runtime-to-runtime validation across live N01-N06 instances remains pending until peers expose compatible endpoints and credentials simultaneously.
- N07 is not to be declared fully fused until all six peer directions have passing discovery, capability invocation, correlation, authentication, timeout, retry and failure-recovery tests.
- Native accelerator execution remains backend-dependent; hardware discovery alone is not treated as execution capability.

## WHAT_NEXT_AGENT_SHOULD_DO

1. Re-read the live N01-N06 Mesh contracts before changing N07 routing or envelope fields.
2. Preserve the canonical `correlationId`, source, target, capability and authentication semantics.
3. When a peer changes its contract, adapt through a compatibility boundary instead of breaking existing consumers.
4. Validate every change with repository CI and inspect the actual tested commit SHA before calling it green.
5. Complete the cross-nucleus matrix and only then finalize N07 fusion and distributed Super GPU scheduling.

## STATUS

N07: IN PROGRESS
Structural Mesh boundary: IMPLEMENTED
CI: CONTINUOUSLY VALIDATED
Full six-nucleus runtime federation: PENDING
Final N01+N06+N07 fusion: PENDING
