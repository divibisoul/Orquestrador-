
# SOUL Operations Runbook

## 1. Golden rules

PRESERVE before AUDIT.
Never delete a working path to make CI green.
Never overwrite a newer concurrent front.
Never fabricate runtime evidence.
Prefer runtime truth over file presence.

## 2. Health incident

1. check readiness;
2. check N07 federation;
3. check target nucleus health;
4. inspect capability discovery;
5. inspect authorization;
6. inspect circuit state;
7. trace by correlationId;
8. compare exact source revisions;
9. reproduce with non-destructive native capability;
10. classify evidence state.

## 3. Mesh security incident

Inspect:
- HMAC failures
- nonce reuse
- timestamp skew
- mTLS identity mismatch
- source spoofing
- rate spikes

Rotate secret if compromise is plausible, using dual-key overlap and E2E verification.

## 4. N07 overload

Inspect worker count, inflight tasks, queue depth, route score and provider latency.
Throttle lower-priority work first.
Never remove cancellation or bounds to recover throughput.

## 5. SARA trace incident

If trace persistence fails:
- mark execution FAIL/BLOCKED;
- preserve snapshot;
- recover DecisionTrace;
- reconcile TemporalVectorDB references;
- reconcile ProvenanceTracker records;
- retry only after persistence is healthy.

## 6. Provider degradation

Provider states must progress through:
DECLARED → CONFIGURED → AVAILABLE → EXECUTABLE.

Missing model weights, binaries or real transaction keeps the provider DEGRADED/BLOCKED.

## 7. Release rollback

Rollback to immutable image digest.
Re-run smoke and provenance checks.
Only then reopen traffic.

## 8. Evidence bundle

Collect:
exact source SHAs, CI run IDs, image digest, schema report, E2E report, trace IDs, SARA cycle trace IDs, failure-injection output and rollback result.

## 9. Engineering cycle

~~~text
PRESERVE
AUDIT
MAP
CORRECT
COMPLETE
CONNECT
CROSS
FUSE
OPTIMIZE
VALIDATE
DOCUMENT
RE-AUDIT
~~~
