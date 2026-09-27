# Jev decision capability in N07

Jev is integrated as an external server-side decision capability. N07 keeps orchestration, authentication, policy, thresholds and final actions.

Configuration:
- JEV_API_KEY: required server-side API key.
- JEV_API_BASE_URL: optional base URL; defaults to https://api.typesafe.ai.
- JEV_MODEL: optional model alias; defaults to jev-latest.

Operation: jev.systemone@1.0.0

Local execution metadata:
- state: decision context.
- questions_json: JSON object containing Jev typed questions.

The structured response is returned in metadata.decision_json. The API key is never returned.

The operation is registered only when JEV_API_KEY is configured, so unavailable external capability is not advertised.

Never place JEV_API_KEY in browser code, Android assets, Git, logs or committed .env files.
