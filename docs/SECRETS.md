# N07 Secrets

N07 never invents or silently substitutes external credentials. Missing credentials put the service in `degraded` / `restricted` mode and are exposed by `/health`.

| Secret | Purpose | Where to obtain |
|---|---|---|
| `N07_APP_TOKEN` | Protects authenticated N07 control and execution endpoints. | Generate a long random application token and store the same value in the deployment environment and GitHub Actions secret. |
| `WEB3_STORAGE_TOKEN` | Authenticates the configured external IPFS/Web3 Storage upload service. | Create an API token in the account used for the project's Web3 Storage service. |
| `SUPABASE_URL` | Identifies the Supabase project used for N07 persistence. | Supabase project settings → API → Project URL. |
| `SUPABASE_SERVICE_ROLE_KEY` | Server-side credential for N07 persistence tables. | Supabase project settings → API → service role key. Keep this value server-side only. |

## GitHub Actions

Configure the four values as repository or environment secrets before running `configured-staging` for credentialed staging:

```text
N07_APP_TOKEN
WEB3_STORAGE_TOKEN
SUPABASE_URL
SUPABASE_SERVICE_ROLE_KEY
```

Do not commit secret values to Git. Do not put them in `.env.example`. The commissioning workflow checks presence and fails explicitly when a required value is missing.

## Restricted mode

With one or more credentials absent, N07 still starts so local health, execution and diagnostics remain available. `/health` reports:

```json
{
  "status": "degraded",
  "secret_mode": "restricted",
  "missing_secrets": ["..."]
}
```

Restricted staging must not attempt external storage or Supabase persistence. The simulated-staging workflow intentionally uses only a locally generated ephemeral token and the local N07 runtime.
