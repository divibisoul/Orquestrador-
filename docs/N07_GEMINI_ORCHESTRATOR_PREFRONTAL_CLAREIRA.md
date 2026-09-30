# N07 Gemini → Orquestrador → Prefrontal → Clareira

## Boundary

N02 remains the canonical owner of the Google Gemini provider capabilities. N07 does not duplicate the provider or claim ownership of the leaf capability IDs.

The N07 Orchestrator exposes explicit delegated operations:

- `gemini.delegate.text@1.0.0`
- `gemini.delegate.multimodal@1.0.0`
- `gemini.delegate.audio.transcribe@1.0.0`
- `gemini.delegate.audio.analyze@1.0.0`
- `gemini.delegate.speech.synthesize@1.0.0`

Each operation is an orchestration boundary. It preserves the canonical N02 capability ID when crossing the Mesh.

## Prefrontal path

For a Gemini request, N07 requires an explicit `candidate_json` describing the action ID, cost, risk, urgency and impact.

The neural input is resolved in this order:

1. the protocol payload;
2. `neural_input_json`;
3. the semantic input through N02's existing `neural.bnc_v2` capability.

The resulting signal is passed to the existing N07 `PrefrontalNeocortex`, which uses the existing N07 neural network and Cortex policy. The decision is committed before the Gemini provider operation is delegated.

No second prefrontal runtime is created.

## Clareira path

The existing N07 Clareira reporter is reused. Gemini lifecycle events are emitted as `StateReport` packets to N01 `clareira.ingest` through the canonical Soul Mesh.

Lifecycle phases are:

- `started`
- `completed`
- `failed`

The original `correlationId` is preserved across BNCv2, prefrontal admission, N02 Gemini execution and Clareira reporting.

No second event bus is created. Vagus/SARA remains the transversal control path supplied by the existing N07 front.

## Evidence boundary

Presence of the gateway code proves structural integration only. Provider availability, N01 availability and external end-to-end execution remain runtime concerns and must be marked accordingly when the corresponding infrastructure is not configured.

Malformed structured Gemini content is rejected explicitly. It is never silently converted into another request shape.

