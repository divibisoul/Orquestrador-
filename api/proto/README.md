# Protobuf contracts

The three service contracts are intentionally small at the wire level. Domain implementations remain in `core/`; generated transport bindings should be produced during build when protobuf tooling is available.
