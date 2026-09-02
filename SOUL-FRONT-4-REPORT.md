# SOUL Front 4 — Coordination Report

## N05

- Bridge contract pinned to `soul-mesh/1`, contract `1.1.0`, bridge `1.1.0`.
- `SOUL-BRIDGE-VERSION.md` mirrors the N07 canonical marker.
- N05 CI validates the canonical N07 marker before the HMAC tests.
- `scripts/runner-diagnostic.sh` captures runner state and can be executed outside GitHub Actions.
- A separate `runner-watchdog` job inspects the bridge job through the GitHub API, records `.diagnostics/runner-failure.json`, uploads diagnostics, and requests one rerun when the failed job has zero provisioned steps.
- The original `FIRST EXECUTION STEP - capture runner state` remains mandatory.
- A pre-start runner failure cannot execute a step inside the failed job; the separate watchdog is therefore the recovery boundary.

## N07

- Missing external credentials are explicit state, not simulated credentials.
- Startup audits `N07_APP_TOKEN`, `WEB3_STORAGE_TOKEN`, `SUPABASE_URL`, and `SUPABASE_SERVICE_ROLE_KEY`.
- Missing credentials produce `status: degraded` and `secret_mode: restricted` on `/health` and are logged at startup.
- N07 remains usable for local health, execution and diagnostics while external storage/persistence remain restricted.
- `staging-simulado` builds the current N07 source, generates an ephemeral local token, verifies degraded health, authenticated status and neural execution, and never configures Web3 Storage or Supabase.
- `configured-staging` remains a hard gate and fails when real secrets are absent.

## Coordination rule

Front 4 does not overwrite N05 PR #13 or N07 PR #21. This work continues from their current heads on isolated `frente-4-*` branches. No existing module or test was deleted, and no failure was converted into a false success.

## Current external blocker

The four real secrets remain pending. Credentialed external deployment and external storage/Supabase round-trip cannot be claimed until those values are configured in GitHub Actions.
