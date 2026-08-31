# State Fabric

Control-plane state is separated from cognitive logic. The current foundation supplies an in-process LRU cache and a Raft integration boundary; production persistence and consensus are explicit adapters.
