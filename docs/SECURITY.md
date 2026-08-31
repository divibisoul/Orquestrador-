# Security

## Current security boundary

The recovery baseline enforces basic execution policy through `security.Validate` and exposes explicit `Authorize` and `AuditRecord` boundaries. These are not a substitute for production identity or durable audit infrastructure.

Required flow:

```text
request -> authentication -> authorization -> policy -> execution -> audit
```

An empty principal is rejected by the authorization extension point. Cost and confidence inputs are range-validated before policy decisions.

## Production requirements still pending

- mTLS and workload identity (for example SPIFFE/SPIRE);
- durable, tamper-evident audit storage;
- least-privilege sandboxing for generated code;
- secret redaction in logs;
- authenticated provider/API boundaries;
- security regression tests and dependency scanning.

## Gemini adapter

The experimental Gemini adapter must not be exposed to an untrusted network until authentication is enforced, credentials are loaded lazily, and internal exception details are removed from HTTP responses.

## Secrets

Never commit provider keys. CI must use repository/environment secret stores. `.env.example` may contain variable names but never real credentials.
