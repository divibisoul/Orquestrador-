# N07 content-addressed storage

N07 exposes a provider-neutral storage capability while using the current Storacha/w3up architecture when it is configured. The legacy Web3.Storage HTTP path remains available only as an explicit compatibility mode.

## Runtime configuration

Preferred production configuration:

```text
STORACHA_MODE=storacha
STORACHA_GUPPY_BIN=/usr/local/bin/guppy
STORACHA_SPACE=did:key:...
STORACHA_DATA_DIR=/var/lib/n07/storacha
STORACHA_GATEWAY_URL=https://storacha.link/ipfs
STORACHA_TIMEOUT=30s
N07_MAX_UPLOAD_BYTES=104857600
N07_APP_TOKEN=<N07 application bearer token>
```

Guppy uses a local agent and UCAN authorization state. The `STORACHA_DATA_DIR` therefore must be persistent in production and must contain an authorized Guppy identity/space before uploads are attempted. Credentials and UCAN material are deployment secrets and must never be committed to Git.

The current Storacha tooling documents `guppy upload source add <space> <path>` followed by `guppy upload <space>`; N07 uses those same commands behind the provider boundary. Storacha calculates the content identifier locally and uploads through the w3up/UCAN service. citeturn0search2turn0search0

## Provider boundary

The backend provider implements:

- bounded uploads with byte-count enforcement;
- safe temporary staging for modern Storacha uploads;
- filename sanitization before staging;
- context cancellation and configurable HTTP/command timeout propagation;
- CID validation before returning or constructing gateway URLs;
- modern Storacha uploads through Guppy;
- legacy Web3.Storage-compatible raw HTTP uploads when explicitly selected;
- CID status through the Storacha gateway in modern mode;
- deterministic IPFS gateway URL generation.

N07 HTTP endpoints are:

```text
POST /v1/storage/upload
GET  /v1/storage/status/<CID>
GET  /v1/storage/object/<CID>
```

All N07 routes are protected by the application bearer token.

## Provider selection

`STORACHA_MODE=auto` selects Storacha only when both `STORACHA_SPACE` and `STORACHA_GUPPY_BIN` are configured. `STORACHA_MODE=storacha` forces the current provider. `STORACHA_MODE=legacy` or `web3.storage` forces the compatibility path.

The legacy endpoint is retained because the repository already contains a provider boundary, but the project should not treat `https://api.web3.storage` as the preferred production integration. The active Storacha implementation is based on the w3up/UCAN architecture and its current Go client/tooling. citeturn0search0turn0search2

## Security and data handling

Data uploaded to w3up/Storacha is publicly retrievable when its CID is known, and the network is designed for content-addressed persistence. Private or sensitive material must therefore be encrypted before upload. citeturn0search1turn0search0

N07 never logs provider credentials. The application token protects N07's control plane; Storacha authorization remains in the provider's agent state.

## Operational commissioning

A configured provider is not proof of a live connection. Structural tests can prove request construction and error handling, but real commissioning requires an authorized Storacha space and a live upload that returns a CID. The repository intentionally does not fabricate that evidence when the deployment environment does not provide the required authorization state.
