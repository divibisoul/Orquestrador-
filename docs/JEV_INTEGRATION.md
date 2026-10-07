# Jev decision capability in N07

Jev is integrated as an external server-side decision capability. The user's jev-api repository is preserved as the Jev website/documentation fork; it is not copied into a nucleus runtime. The executable model is called through Jev's documented HTTP API. N07 keeps orchestration, authentication, policy, thresholds and final actions.

Configuration:
- JEV_API_KEY: required server-side API key.
- JEV_API_BASE_URL: optional base URL; defaults to https://thejevai.com.
- JEV_MODEL: optional model alias; defaults to jev-latest.
- Endpoint: POST /v1/systemone.

Operation: jev.systemone@1.0.0

Local execution metadata:
- state: decision context.
- questions_json: JSON object containing Jev typed questions.

The structured response is returned in metadata.decision_json. The API key is never returned.

The executable operation is registered only when JEV_API_KEY is configured. The N07 identity/topology names the local decision-specialist agent; runtime capability discovery exposes Jev as executable only when the server-side key is present.

Never place JEV_API_KEY in browser code, Android assets, Git, logs or committed .env files.

<!-- Provenance: divibisoul/jev-api main supplies the current published Jev API contract/documentation; N07 owns the adapter and action boundary. -->
