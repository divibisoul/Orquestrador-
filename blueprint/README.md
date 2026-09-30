# ATLAS OMEGA FLUX 8X — Functional Affinity Layer

This package is an additive integration layer for the existing SOUL architecture.

It does not create a second Mesh, second capability authority, fake implementation,
or new ownership model. It maps the blueprint's eight functional families onto
capabilities that already exist in the repositories inspected on 2026-09-29.

## Rule

The number in the blueprint is a functional family, not a repository destination.

The resolver answers which existing capability has the strongest functional affinity,
which supporting capabilities can participate without taking ownership, which
implementations are preserved, and which areas remain structurally incomplete.

## Evidence boundary

A registry entry is evidence of a mapped implementation surface only. It is not runtime
proof. Online promotion still requires the repository's normal
DISCOVER -> NEGOTIATE -> EXECUTE -> CORRELATE -> VERIFY evidence chain.

The learning family intentionally does not declare 20/20 paradigms complete. The
inspected source evidence does not justify that claim yet, so the gap remains explicit.

## Non-destructive invariant

ValidateManifest rejects duplicate canonical owners for one leaf capability while
allowing the same functional family to contain complementary capabilities owned by
different nuclei.
