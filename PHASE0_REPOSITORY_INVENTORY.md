# SOUL Phase 0 — Repository Inventory

Inventory reconciled against GitHub repository state and recent GitHub Actions evidence on 2026-09-28.

| Unit | Repository | Mesh contract | Lockfile state | Current CI evidence |
|---|---|---|---|---|
| N01 | `aeternum-core-29` | `soul-mesh/1`, `1.1.0` | `package-lock.json` present; no pnpm lock | Latest validation/build checks observed successful. |
| N02 | `Eternium-` | `soul-mesh/1`, `1.1.0` | No `package-lock.json` and no `pnpm-lock.yaml` in current root inventory | Latest validation/forensics checks observed successful. |
| N03 | `nexus-aeternum-fusion` | `soul-mesh/1`, `1.1.0` | `package-lock.json` present; no pnpm lock | Dependency-resolution and dependency-review runs currently failing; job-step logs were not exposed by GitHub connector, so root cause remains unmeasured. |
| N04 | `nextjs-ai-chatbots` | `soul-mesh/1`, `1.1.0` | `pnpm-lock.yaml` present; no package lock | Latest CI and Soul Mesh checks observed successful. |
| N05 | `nextjs-ai-chatbot` | `soul-mesh/1`, `1.1.0` | `pnpm-lock.yaml` present; no package lock | Latest Mesh/bridge checks observed successful. |
| N06 | `nextjs-ai-chatbot-2000` | `soul-mesh/1`, `1.1.0` | `pnpm-lock.yaml` present; no package lock | Latest Mesh/authority/validation checks observed successful. |
| N07 | `Orquestrador-` | `soul-mesh/1`, contract `1.1.0` | `go.sum` present (empty in current root inventory); Go module uses `go 1.25` | Current Go verification path is green; external staging remains configuration-gated. |
| SARA | `SARA` | Independent SARA service; Mesh integration preserves correlation and authority boundary | Python packaging is driven by `pyproject.toml`; no committed `requirements.txt`, `poetry.lock` or `uv.lock` found in the checked root paths | Latest CI/self-check observed successful. |

## Workflow evidence

### N01
Current repository contains Mesh/validation/build workflows. Latest observed validation and build runs were successful.

### N02
Current repository contains validation/diagnostic Mesh workflows. Latest observed validation/forensics runs were successful.

### N03
The current branch history shows successful bridge/lock-repair work, but the latest dependency-resolution and dependency-review runs failed. The connector did not return job-step records or logs for those runs. This is **INCOMPLETO**, not a permission to guess dependency changes.

### N04
Current CI uses pnpm with frozen lockfile validation and runs the N04 test suite/typecheck/build.

### N05
Current CI uses pnpm/frozen lockfile and validates Mesh, SARA and type/build surfaces.

### N06
Current CI uses pnpm/frozen lockfile plus N06 processor, authority, Mesh HMAC and channel-contract tests.

### N07
Current CI validates architectural/source integrity, Go vet, tests, race tests, build and container integrity. The real peer-federation test is now an explicit integration gate rather than a default unit-test substitute.

### SARA
Current CI installs the package with dev dependencies, runs the deterministic SARA self-check, then pytest.

## Phase 0 decisions

- **REAL:** contract alignment source evidence exists across the seven Mesh nuclei.
- **REAL:** N01, N02, N04, N05, N06, N07 and SARA currently have successful recent CI evidence.
- **INCOMPLETO:** N03 dependency-resolution/dependency-review failure lacks persisted step logs required for a safe root-cause patch.
- **BLOCKED:** N07 external peer commissioning because live Mesh URLs/HMAC secrets are not available in the current GitHub runner configuration.
- **NOT MEASURED:** full seven-runtime simultaneous federation.

No package manager is being normalized across repositories in this phase; that would be a broad architectural change without a demonstrated CI blocker and would violate the minimum-diff rule.
