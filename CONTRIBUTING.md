# Contributing

## Development rules

1. Work from a feature or recovery branch; never push experimental changes directly to `main`.
2. Preserve existing behavior unless the change is documented in `ISSUES.md`.
3. Keep `core/*` as the canonical domain layer. Infrastructure adapters must not silently fork domain logic.
4. New interfaces must distinguish implemented behavior from stubs/boundaries.
5. New external effects require an explicit security-policy decision and tests.
6. Update tests and documentation in the same change.
7. Never commit secrets.
8. Keep traceability from task → decision → delegation → execution → outcome.

## Required checks

```bash
go fmt ./...
go vet ./...
go build ./...
go test ./...
go test ./... -race
```

For changes involving security, mesh, state, providers or public contracts, add focused tests and explain compatibility impact in the pull request.

## Pull requests

Describe the problem and evidence, files/packages changed, compatibility impact, tests executed and their output, known stubs, and any migration/deprecation plan. Do not merge until required CI checks are green.
