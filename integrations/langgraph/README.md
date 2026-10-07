# SOUL N07 LangGraph bridge

This is an optional planner/state adapter. It does not become another SOUL nucleus, another Mesh, or another public API surface.

The graph is intentionally minimal:

decompose -> route -> aggregate

The route node sends requests to the existing N07 /v1/chat/completions surface. N07 remains the only OpenAI-compatible ingress, and its implementation uses the already-existing CallBestDynamic peer router for ai.generate.

Install:

python -m pip install -e integrations/langgraph

Run:

printf '%s\n' '{"tasks":[{"prompt":"hello"}]}' | python integrations/langgraph/run.py

Required environment for a real route:
- SOUL_N07_URL
- N07_APP_TOKEN

This bridge does not assert online readiness unless a real N07 endpoint and configured peer are available.

LangGraph 1.2.12 is pinned because that is the current PyPI release observed on 2026-09-27. LangGraph's documented Python API uses StateGraph, START, END, and compile() for executable graphs.
