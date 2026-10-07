
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

The current N07 repository pins 16 upstream repositories. Their implementation must flow through:
provenance → contract → adapter → Mesh route → controlled execution → verification → observability → re-audit.

Existing provider pins are preserved.

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

