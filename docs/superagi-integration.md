# SuperAGI integration

Source audited: TransformerOptimus/SuperAGI.

## External inventory

The upstream project currently identifies itself as an open-source autonomous-agent framework. The current GitHub release is v0.0.14. The project is MIT licensed.

The upstream SuperAGI-Tools repository currently contains community toolkits including:
- duck_duck_go
- google_analytics
- news_api
- notion

The official Node client exists as TransformerOptimus/SuperAGI-Node-Client.

## SOUL integration strategy

SOUL/N07 does not copy the SuperAGI implementation. It exposes an adapter at integrations/superagi/client.go.

The bridge is fail-closed:
- SUPERAGI_BASE_URL is mandatory.
- SUPERAGI_API_KEY is mandatory.
- no credentials are embedded in source.
- network execution uses the existing Go HTTP stack.
- create-agent path defaults to /agents/create.
- agent-run path defaults to /agents/run.
- path configuration remains explicit in the Client structure for compatibility with a deployment whose API routes differ.

## Toolkits

Toolkits are catalogued as external capabilities, not silently copied into SOUL. Any future native adaptation must preserve upstream attribution and the applicable MIT license text.

## Runtime status

The integration becomes EXECUTABLE only when a live SuperAGI endpoint and API key are supplied. Without those external runtime inputs the bridge must report BLOCKED, not fabricate connectivity.

## Evidence

This PR contains the source adapter and configuration/build validation. It intentionally does not claim a live SuperAGI network run without a configured endpoint.
