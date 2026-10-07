# N07 Neural Federation

N07 owns the canonical neural execution engine. N01-N06 consume it through the canonical Soul Mesh rather than copying the model implementation.

## Runtime path

Nucleus → N07 Mesh → `neural.forward@1.0.0` / `neural.learn@1.0.0` → neural runtime → correlated response.

## Security

Requests use Mesh contract `1.1.0`, HMAC-SHA256, source-bound nonce, correlation ID, timestamp validation, bounded payloads and request cancellation.

## Local adapters

Every N01-N06 repository receives a native TypeScript `N07NeuralBridge` adapter. The adapter validates finite numeric input, signs the exact Mesh envelope shape, propagates correlation, enforces a timeout and validates response contract/correlation before returning data.

## Ownership

N01-N06 retain ownership of their native processors, tools and capabilities. N07 supplies the shared neural orchestration/processing service. Existing local neural logic may remain where it provides nucleus-specific preprocessing, but delegation to N07 is the canonical cross-nucleus path for shared neural processing.

## Failure rule

A missing N07 endpoint, invalid secret, contract mismatch, timeout or remote error is returned explicitly. No success value is fabricated and no alternate nucleus is silently substituted.

## Cross-front handoff

This change is additive and intentionally isolated from concurrent federation branches. Before modifying shared Mesh files, reread their current GitHub SHA. On a SHA conflict, fetch the newest file, preserve its unrelated work, and reapply only the necessary change.
