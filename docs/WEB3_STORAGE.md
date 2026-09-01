# Web3 Storage integration

N07 exposes Web3 Storage as a provider capability without embedding credentials or coupling the Mesh to a storage vendor.

## Runtime configuration

Set these values in the deployment environment:

```text
WEB3_STORAGE_TOKEN=<provider credential>
WEB3_STORAGE_API_URL=https://api.web3.storage
WEB3_STORAGE_TIMEOUT=60s
WEB3_STORAGE_MAX_UPLOAD_BYTES=104857600
N07_APP_TOKEN=<N07 application bearer token>
```

`WEB3_STORAGE_TOKEN` is never committed to Git. The adapter returns `ErrNotConfigured` when it is absent, allowing N07 to start without falsely claiming provider connectivity.

## Provider boundary

The adapter implements:

- upload → CID;
- CID status;
- authenticated upload listing / health check;
- bounded request size;
- context cancellation and timeout;
- provider error propagation;
- configurable API endpoint for controlled migration to another compatible endpoint.

N07 routes the capability surface as:

```text
storage.web3.upload@1.0.0
storage.web3.status@1.0.0
```

HTTP control endpoints are protected by `N07_APP_TOKEN`:

```text
POST /storage/web3/upload
GET  /storage/web3/status?cid=<CID>
GET  /storage/web3/health
```

The upload endpoint accepts a multipart field named `file` and returns the resulting CID.

## Security

Web3 Storage content is content-addressed and retrievable through IPFS. Do not upload unencrypted private or sensitive information. Provider credentials remain deployment secrets. N07 only exposes the resulting CID and provider status to authenticated application callers.

## Operational validation

A configured provider is not considered connected merely because an environment variable exists. `/storage/web3/health` performs an authenticated provider request. Successful upload is proven only when the provider returns a CID.

The implementation uses the documented Web3 Storage HTTP API boundary. Web3 Storage has also introduced the newer w3up/UCAN architecture; `WEB3_STORAGE_API_URL` keeps the provider boundary replaceable so migration can occur without changing Mesh or N07 capability contracts.
