# SOUL — SuperAGI Control Fabric

The SOUL SuperAGI Fabric is a composition of already-owned native capabilities. It is not a second nucleus and is not a scientific claim of general intelligence.

Architecture:

request
-> canonical Mesh (N07)
-> resident multi-agent boundary
-> explicit provider adapter
-> accelerator admission
-> SuperGPU lease
-> bounded compute / Octacore
-> Prefrontal/JEV/SARA validation and evidence
-> result with the same correlation context.

Multi-agent layer:
- Multi-Agent Facade remains the canonical N07 facade for CrewAI.
- Agent Arsenal is the bounded sidecar for Superpowers, ECC and Ruflo.
- External providers remain pinned implementation sources.

GPU layer:
- N07 exposes mesh.supergpu.describe@1.0.0.
- N07 exposes mesh.supergpu.execute@1.0.0.
- N07 exposes mesh.supergpu.parallel@1.0.0.
- Accelerated routes require a real available accelerator.
- CPU compatibility can be requested explicitly and is marked accelerator=false; it is never accepted as GPU evidence.

Fused execution:
superagi.fabric.execute@1.0.0 composes an agent stage with an accelerator stage while preserving the original correlation ID.

Whole-system distribution:
N01, N02, N03, N04, N05, N06, SARA and JEV are registered as Mesh consumers of the fused N07 fabric. Their native ownership is unchanged.

Truth states:
VERIFIED_BY_CI means code, contracts and tests passed.
ENVIRONMENT_DEPENDENT means actual accelerator hardware may or may not be present.
ENVIRONMENT_AND_CREDENTIAL_DEPENDENT means external agents require their real runtime and credentials.
NOT_ASSERTED means no claim is made that the composition itself proves general intelligence.
