# SOUL Engineering Execution Policy

This is the standing engineering policy for every future GitHub modification across N01-N07.

## 1. Truth before narration

Never infer system state from prior claims, task text, percentages, filenames, or expected architecture. Before changing code, inspect the current branch, current HEAD, relevant files, recent commits from parallel fronts, and the latest applicable CI/runtime evidence.

## 2. Repair before reporting

When a defect, inactive area, inconsistency, incomplete capability, failing test, broken integration, stale contract, or false-positive validation is found, the default action is to repair it in the current state. Do not stop because a failure is severe. Diagnose the failure, research viable alternatives when necessary, implement the safest concrete correction, then validate again.

## 3. External research is an engineering tool

When the repository, compiler, runtime, platform, protocol, dependency, or CI environment blocks progress, use authoritative external documentation or current technical sources to identify alternative solutions. Do not treat an unfamiliar failure as a stopping condition.

## 4. Parallel-front reconciliation

Before creating a new implementation, inspect recent work from other fronts. Preserve correct concurrent changes, adapt to their current interfaces, and avoid duplicate implementations, duplicate owners, duplicated registries, or conflicting protocol definitions.

## 5. Write -> read back -> execute -> re-check

Every substantive code change must follow this evidence chain:

1. Write the correction to GitHub.
2. Read the changed file back from the resulting revision.
3. Run or trigger the narrowest meaningful test.
4. Run broader CI where required.
5. Inspect the resulting evidence against the exact commit that was changed.
6. If the state moved underneath the change, reconcile again instead of assuming the result still applies.

## 6. Runtime truth

A build proves compilation. A unit test proves the tested behavior. An integration test proves the exercised boundary. A real end-to-end transaction is required before marking a cross-nucleus communication path as operational.

## 7. Time and delivery dashboard

Every long-running audit must track elapsed audit time, current revision, current state, validation evidence, remaining blockers, and the next executable correction. Use a compact progress graph when reporting substantial work. Do not use percentage completion as evidence of correctness.

## 8. Failure does not terminate the audit

A failing command is input to diagnosis, not the end of the task. Retry with the corrected hypothesis, change approach when appropriate, reduce scope to isolate the failure, or use another supported implementation path.

## 9. Ownership and composition

Each executable capability has one authoritative owner. Other nuclei delegate through the canonical Mesh contract rather than copying specialized runtimes. New functions should emerge from safe composition of existing capabilities where useful, not from parallel duplicate implementations.

## 10. N07 parallel development, final activation

N07 is engineered continuously with the other nuclei: protocol, security, observability, telemetry, orchestration, compute, tests, adapters, and performance all remain active work. Final N07 activation/fusion remains gated until the complete N01-N06 ingress/egress, capability, tool, authorization, correlation, error, retry, resource, and shutdown surfaces are inventoried and validated.

## 11. Delivery gate

Do not declare the system online until the relevant code is present on the claimed revision and the required validation evidence is green for that same revision. When evidence is incomplete, report the exact unresolved state instead of upgrading it rhetorically.
