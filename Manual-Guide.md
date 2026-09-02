# N07 Manual Guide — Storacha / Web3 Storage

## Storage state model

N07 exposes Storage as an explicit state. When no authorized Storacha credential is configured, the runtime reports `status: not-configured`; this is neither a successful upload nor a failed provider operation.

The health/capability response includes a `cureSuggestion` explaining that an authorized Storacha credential and Space must be configured before commissioning.

## Required configuration

Modern Storacha/Guppy deployments require an authorized Space and local UCAN/agent state. Configure the runtime with:

- `STORACHA_UCAN` or `WEB3_STORAGE_TOKEN` for the authorized credential material used by the deployment.
- `STORACHA_SPACE` for the target Space.
- `STORACHA_GUPPY_BIN` when `guppy` is not on `PATH`.
- `STORACHA_DATA_DIR` when persistent local Storacha agent state must live at a controlled path.
- `STORACHA_TIMEOUT` and `N07_MAX_UPLOAD_BYTES` as operational limits.

Never commit credentials, place tokens in source code, or use fabricated values in CI.

## Commissioning sequence

1. Obtain an authorized Storacha Space and UCAN through the official Storacha account/Space workflow.
2. Configure the credential and Space as protected deployment secrets/environment variables.
3. Restart N07 so the provider state is reconstructed from the configured environment/local agent state.
4. Confirm `/v1/health` reports Storage as `ready` only when the provider is actually configured.
5. Execute the real upload commissioning test. The test must return a provider-issued CID and the exact number of bytes uploaded.
6. Verify the CID through the status/gateway path.

Without credential material, the commissioning workflow intentionally logs:

`SKIPPED: Storage not configured. Set STORACHA_UCAN or WEB3_STORAGE_TOKEN to enable real upload tests.`

This state must not be reported as a green real-upload commissioning result.

## Legacy web3.storage compatibility

The legacy HTTP adapter remains isolated for compatibility. New deployments should prefer the modern Storacha path because the historical web3.storage upload API is deprecated for new uploads.
