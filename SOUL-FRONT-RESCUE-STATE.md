# SOUL — FRONT RESCUE STATE
Date: 2026-09-30

This file is the cross-chat checkpoint for the repository archaeology/rescue cycle.
It exists so a parallel engineering front can resume from repository truth instead of replaying the same historical PR.

## Terminal-state rules
A historical front must end in one of:
- RESCUED: unique functionality was adapted into current MAIN.
- ABSORBED: the current MAIN already contains the useful functionality.
- BLOCKED_ENV: the code path exists but external endpoint/credential/infrastructure evidence is missing.
- UNMEASURABLE: the repository does not currently provide enough evidence to distinguish states.

Never reopen a terminal front only because its old branch is large. Compare the exact file/function first.
Never merge a historical PR wholesale when current MAIN has a newer authority.

## Rescued / activated in this cycle

N01 — aeternum-core-29
- #75 merged 2a49ff97bca4620b2ae2a65f9de771efde55805c: recovered Aeternum Mesh bridge/capabilities.
- #76 merged 4bd0ba59533117c6c33fdda8fa0a22089ff25992: HortaCore vascular state + Mesh traffic observer/bridge.
- #60/#63/#65/#51/#66: historical functionality compared; #60/#51/#66 absorbed, #63/#65 selectively rescued and closed.
- Remaining proof state: source integration is present; real deployed Mesh/credential evidence is still not claimed.

N02 — Eternium-
- #34 merged f1a446f84f1350771ab2d3cc613626dc3a9d178b: perception input, inference.reason, SuperGPU helpers, SARA correlation/payload guard and fail-closed executor hook.
- #35 merged a8eba2f2d4cfbd27913bd469512401ffb0ac66bc: Clareira event/packet contract + drop/uptime bookkeeping.
- Historical #19/#20 are terminally rescued/closed.
- Remaining proof state: source integration is present; deployed N02↔N01/N07 evidence requires real endpoints/credentials.

N03 — nexus-aeternum-fusion
- #27 merged dd629ebcef21595a165cab46dc919328950350ad: neural.parameters on canonical N07 bridge.
- #28 merged 81c90ddd7eadcec1c894ca333359f9be3c2e1c4a: PerceptionModule, Clareira bridge, SuperGPU helpers, N02 reasoning path, N07 cognitive HMAC bridge, N07 peer validation.
- Historical #14/#15/#16 closed after selective rescue.
- Remaining proof state: source integration is present; live cross-nucleus commissioning is UNVERIFIED/BLOCKED_ENV.

N04 — nextjs-ai-chatbots
- #34 merged 5ecbb12872f328933e2a018356ac3a754bd94663: composition capabilities advertised in discovery.
- #36 merged 5205852aa275fc96d56e26e7ccb85100c705cd13: neural.parameters consumed without dropping learning.feedback metadata.
- #29 historical composition PR is ABSORBED and closed.
- Remaining proof state: source integration is present; live N04→N07 transaction is not claimed.

N05 — nextjs-ai-chatbot
- neural.parameters rescue merged in #34: db9908ef2cbec208bbb70e4beeb3745c67921eee.
- federated SARA context merged in #35: 5b3183f58f7be6d76af8beb4eae7cbddf3755984.
- RGO finding boundary merged in #36: 4269ef108ac6e1beae725ad95d566ff0ba3586ff.
- Historical #27/#32 are terminally rescued/absorbed and closed.
- Remaining proof state: source integration present; real SARA/N07 online verification is not claimed.

N06 — nextjs-ai-chatbot-2000
- #27 merged f7668a9892c734e25f6a229dca3a74e7a38b7ae1: peer route + authenticated handshake.
- #28 merged 3f6bc3936d890220fa562b12b88e153e03d2e1b1: neural.parameters bridge consumption.
- #11/#24/#25 historical fronts are ABSORBED/closed where current MAIN already contains their functionality.
- Remaining proof state: Mesh reachability is environment-gated.

N07 — Orquestrador-
- #75 merged 30d6c811d2b28be720b0e77a3d40788fe4ec255a: canonical learning receiver.
- #77 merged 9b3df91cd27ac8c09e1721805389d86721c581c7: cognitive loop/planner.
- #78 merged 547d781509438ac3a61d5d599040cbf391293e3f: primordial composition.
- #80 merged 8dcfa3823d79262ac80234a850e80f2de047b824: prefrontal executive.
- #82 merged 6f165b730a6d881afba24eb1a423928406493fcd: learning→prefrontal observations.
- #83 merged 12a9b4c6ac9299e3a76d7ed25c6d54b61a4469ba: buried prefrontal task/uncertainty semantics.
- #84 merged f38f1fef3fe7800f6e9f696cece17a589faac64f: HortaCore 29-module catalog + semantic memory + pgvector boundary.
- #85 merged 92a2e0c8624b957de16503cb165aff0e122e1c92: real-only federation E2E gate.
- #86 merged 0e8eeaa7a0165f760cbd8d69192a1f9b58a47d: learned route selection/outcome observation.
- Historical #61/#64/#67/#68/#81 closed after selective rescue. #69 remains an archaeology source only; no wholesale merge.
- Important proof rule: current federation E2E is credential-gated and synthetic peer doubles are forbidden. No online proof is claimed without real endpoint/HMAC evidence.

SARA
- #29 merged 730b622ad158a5119a7c7be55bbc1a266081f598: authenticated /v1/hortacore/assess, backed by existing AeternumChimeraBridge.fuse_assessment().
- Existing probabilistic/Bayesian runtime and federated-context infrastructure were verified as already present and were not duplicated.
- Historical #18 is ABSORBED/closed.
- Remaining proof state: external SOUL↔SARA transactions still depend on real deployment/credentials.

Jev
- No open historical PR requires action in this rescue cycle.
- Do not fabricate Jev integration status; inspect its current boundary before claiming execution.

## Immediate next front
Do not start another broad archaeology sweep. Choose ONE remaining open historical PR with the highest verified unique delta, compare its exact files against MAIN, and terminate that front as RESCUED/ABSORBED/BLOCKED_ENV/UNMEASURABLE.

## Cross-chat unblock instruction
A parallel ChatGPT engineering front should:
1. Read this file and its nucleus SOUL-HANDOFF.md.
2. Use the listed merge SHA as the current source of truth.
3. Never reopen/merge a terminal historical PR without a file-level difference.
4. When blocked by missing deployment credentials/endpoints, record BLOCKED_ENV and move to the next bounded task instead of looping.
5. Leave its own handoff with one concrete next action for the next front.

## Evidence discipline
IMPLEMENTED means code is present in current MAIN.
CI_VALIDATED means a real workflow/test run is observed.
ONLINE_VERIFIED means a real deployed endpoint transaction is observed.
BLOCKED_ENV means execution is blocked by missing external environment.
UNMEASURABLE means available repository evidence is insufficient.
These states must never be conflated.
