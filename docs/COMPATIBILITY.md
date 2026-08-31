# Compatibility

SOUL-NEXUS uses capability-based contracts rather than hard-coded model vendors.

- **Soul:** implement a bridge that maps Soul node announcements and task envelopes to Nexus `NodeAnnounce`/`TaskSpec`.
- **LangChain/CrewAI/AutoGPT:** wrap agents as capability providers behind the task boundary.
- **Kubernetes:** deploy control/worker components as services; Kubernetes remains substrate orchestration while Nexus provides cognitive scheduling.
- **Ray/Dask:** expose compute workers through adapters; Nexus decides intent/capability while the backend executes data-parallel work.
- **HTTP/WebSocket/MQTT:** transport adapters may terminate at the Mesh boundary.

Compatibility is achieved through explicit contracts, not by pretending every backend has identical semantics.
