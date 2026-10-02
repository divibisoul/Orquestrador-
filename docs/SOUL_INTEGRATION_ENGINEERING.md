# SOUL Integration Engineering

## Objective

The external repositories are entering SOUL for a functional reason: each covers a capability gap or strengthens an existing capability without replacing the canonical SOUL authority.

The engineering target is:

source -> capability contract -> explicit adapter -> canonical routing -> controlled execution -> evidence -> observability -> re-audit

## Evidence states

- REAL: executed and verified with runtime evidence.
- PROJECTED: contract and adapter path exist, but runtime execution is not yet proven.
- BLOCKED: execution is intentionally prevented because required infrastructure/configuration is absent.
- UNMEASURABLE: the current environment does not provide enough evidence to classify the runtime.

A Git submodule or manifest entry alone is never runtime proof.

## Engineering agents

### Provenance Engineer
Maintains upstream URL, pinned revision, attribution and source identity.

### Capability Adapter Engineer
Converts an upstream capability into a SOUL capability contract without blindly replacing native SOUL code.

### Runtime Integration Engineer
Connects adapters to the existing N01/N02/N03/N04/N05/N06/N07/SARA/JEV boundaries and canonical Mesh.

### Verification Engineer
Builds contract tests, smoke tests and runtime evidence gates.

### Resilience and Observability Engineer
Enforces fail-closed behavior, correlation IDs, bounded execution, sandbox requirements, Mesh integrity and Clareira reporting.

## Mandatory provider record

Every provider must have:

1. reason for inclusion
2. target component
3. capability namespace
4. activation boundary
5. owner/authority
6. adapter state
7. verification contract
8. evidence state
9. failure path
10. re-audit trigger

No provider is executable merely because its repository is present.
