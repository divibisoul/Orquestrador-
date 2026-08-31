# State Fabric / Raft

Contract boundary for strongly consistent control-plane metadata. The production adapter is intentionally isolated so Raft implementations can be supplied without coupling cognitive modules to a storage vendor.

Required properties: leader election, replicated metadata, checkpoints, durable workflow state and recovery.
