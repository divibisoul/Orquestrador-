
# SOUL Capability Registry and Agent Discovery

## 1. Authorities

N07 = cross-nucleus federation/control plane.  
N01–N06 = native capability owners according to capability matrix.  
SARA = native owner of sara.* regenerative/governance operations.

External providers never become authorities merely because they are vendored/submoduled/pinned.

## 2. Native authority map

| Family | Owner | Minimum execution proof |
|---|---|---|
| inference.* | N05 | native handler + live result |
| conversation.* | N05 | native handler + live result |
| document.* | N04 | native handler + live result |
| tool.* | N04 | native handler + live result |
| audio.* | N03 | provider transaction |
| cognition.* | N06 | native cognition transaction |
| agent.* | N02/N06 | exact capability owner transaction |
| mesh.* | N01/N07 | protocol E2E |
| compute./supergpu.* | N07 | bounded parallel execution + validation |
| sara.* | SARA | 12-phase trace linkage |

## 3. SOUL-25 provider fabric

The pinned public-upstream inventory is reconciled across **31 external repositories**:
- 16 primary sources and 6 complementary sources have records in the canonical adapter registry (22 records total). A registry entry is not proof that the provider is enabled or executable.
- 6 expansion sources and 3 GRCE complements are visible through federation discovery as `PROJECTED`; they are intentionally excluded from the runtime-routeable adapter map until a native adapter exists and passes its contract.
- Source pins must match the capability manifest, Git submodule link and checked-out source revision. Unregistered external submodules fail the integrity gate.
- Candidate-only sources that are not materialized as pinned submodules and historically blocked sources remain outside the 31-source route graph.

Every canonical nucleus N01–N07 can discover the same N07 external-federation catalog through the canonical Mesh. The external provider retains its mapped owner, and every request still passes capability authorization, source-pin validation, timeout/deadline policy, HMAC/mTLS and provenance checks. No upstream repository becomes an eighth nucleus or a second control plane.

Required path:
provenance → contract → adapter → Mesh route → controlled execution → verification → observability → re-audit.

Runtime activation remains fail-closed and disabled by default. Only an authenticated, exact-pin, real capability transaction may promote an expansion beyond `PROJECTED`.

## 4. Versioning

Operation identity:
name@major.minor.patch

Explicit version requires exact compatibility.
Unversioned requests may use newest compatible version.
Major-version crossing is forbidden unless an adapter is registered and tested.

## 5. Availability lifecycle

DECLARED = manifest only  
CONFIGURED = source/config exists  
AVAILABLE = handler/provider health passes  
EXECUTABLE = real capability transaction passes  
DEGRADED = provider exists but dependency is not fully available  
BLOCKED = required proof cannot complete

Only EXECUTABLE is routeable as a normal PASS capability.

## 6. TTL

Capability TTL: 60s.
Agent heartbeat: 15s.
Stale: after 3 missed heartbeats.
Positive route cache: 30s.
Negative cache: 5s.

## 7. Route score

score = 0.40 health
      + 0.25 latency
      + 0.20 inverse recent-failure rate
      + 0.10 capacity headroom
      + 0.05 priority fit

Hard filters execute before scoring.

## 8. Agent registration

A registration must identify:
id, nucleus, role, lifecycle, capabilities, tools, resourceLimits, evidenceState, revision and provenance.

Resident agents must obey:
- native-owner rule
- evidence requirement
- no repository writes by default
- human approval for delete/payment/deploy/credential rotation as required by the existing resident-agent contract

## 9. Discovery reconciliation

The CI gate compares:
- native capability registry
- ownership matrices
- external capability manifest
- declared route endpoints
- evidence state
- current source revision

A stale or contradictory registry causes BLOCKED release state.

