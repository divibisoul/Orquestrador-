"""Optional SOUL LangGraph bridge.

Authority boundaries:
- LangGraph is a planner/state graph only.
- N07 remains the sole execution/orchestration authority.
- The route node calls N07 /v1/chat/completions, whose implementation delegates
  inference through N07's existing CallBestDynamic path.
- No Mesh identity, Mesh endpoint, provider, or SARA core is created here.
"""

from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
import uuid
from typing import Any, TypedDict

from langgraph.graph import END, START, StateGraph


MAX_TASKS = 16
MAX_PROMPT_CHARS = 64_000
MAX_RESPONSE_BYTES = 1 << 20


class SoulGraphState(TypedDict, total=False):
    tasks: list[dict[str, Any]]
    routed: list[dict[str, Any]]
    aggregate: dict[str, Any]


def _n07_url() -> str:
    return os.environ.get("SOUL_N07_URL", "http://127.0.0.1:8080").rstrip("/")


def _n07_token() -> str:
    token = os.environ.get("N07_APP_TOKEN", "").strip()
    if not token:
        raise RuntimeError("N07_APP_TOKEN is required for LangGraph execution")
    return token


def _task_prompt(task: dict[str, Any]) -> str:
    prompt = task.get("prompt")
    if isinstance(prompt, str) and prompt.strip():
        value = prompt.strip()
    else:
        messages = task.get("messages")
        if isinstance(messages, list):
            parts: list[str] = []
            for message in messages:
                if not isinstance(message, dict):
                    continue
                role = str(message.get("role", "")).strip()
                content = message.get("content")
                if isinstance(content, str) and content.strip():
                    parts.append(f"[{role or 'user'}]\n{content.strip()}")
            value = "\n\n".join(parts).strip()
        else:
            value = ""
    if not value:
        raise ValueError("each task requires prompt or messages")
    if len(value) > MAX_PROMPT_CHARS:
        raise ValueError("task prompt exceeds configured size limit")
    return value


def decompose(state: SoulGraphState) -> SoulGraphState:
    raw = state.get("tasks")
    if not isinstance(raw, list) or not raw:
        raise ValueError("tasks must be a non-empty list")
    if len(raw) > MAX_TASKS:
        raise ValueError(f"tasks exceeds limit of {MAX_TASKS}")

    normalized: list[dict[str, Any]] = []
    for index, task in enumerate(raw):
        if not isinstance(task, dict):
            raise ValueError(f"task {index} must be an object")
        normalized.append(
            {
                "id": str(task.get("id") or f"task-{index + 1}"),
                "prompt": _task_prompt(task),
                "model": str(task.get("model") or "soul-auto"),
            }
        )
    return {"tasks": normalized, "routed": []}


def _call_n07(task: dict[str, Any]) -> dict[str, Any]:
    payload = {
        "model": task["model"],
        "messages": [{"role": "user", "content": task["prompt"]}],
        "stream": False,
    }
    body = json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode("utf-8")
    request = urllib.request.Request(
        f"{_n07_url()}/v1/chat/completions",
        data=body,
        method="POST",
        headers={
            "Authorization": f"Bearer {_n07_token()}",
            "Content-Type": "application/json",
            "X-Request-ID": str(uuid.uuid4()),
        },
    )
    try:
        with urllib.request.urlopen(
            request,
            timeout=float(os.environ.get("SOUL_N07_TIMEOUT", "60")),
        ) as response:
            raw = response.read(MAX_RESPONSE_BYTES + 1)
            status = int(response.status)
    except urllib.error.HTTPError as exc:
        detail = exc.read(4096).decode("utf-8", errors="replace")
        raise RuntimeError(f"N07 HTTP {exc.code}: {detail}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"N07 transport error: {exc.reason}") from exc

    if len(raw) > MAX_RESPONSE_BYTES:
        raise RuntimeError("N07 response exceeds configured size limit")
    if status < 200 or status >= 300:
        raise RuntimeError(f"N07 HTTP {status}")

    decoded = json.loads(raw.decode("utf-8"))
    choices = decoded.get("choices")
    if not isinstance(choices, list) or not choices:
        raise RuntimeError("N07 response has no choices")
    message = choices[0].get("message") if isinstance(choices[0], dict) else None
    content = message.get("content") if isinstance(message, dict) else None
    if not isinstance(content, str) or not content.strip():
        raise RuntimeError("N07 response has no textual assistant content")

    metadata = decoded.get("metadata")
    return {
        "id": task["id"],
        "status": "routed",
        "text": content,
        "metadata": metadata if isinstance(metadata, dict) else {},
    }


def route(state: SoulGraphState) -> SoulGraphState:
    tasks = state.get("tasks", [])
    return {"routed": [_call_n07(task) for task in tasks]}


def aggregate(state: SoulGraphState) -> SoulGraphState:
    routed = state.get("routed", [])
    if not routed:
        raise ValueError("route produced no results")
    if any(item.get("status") != "routed" for item in routed):
        raise RuntimeError("one or more routed tasks failed")
    return {
        "aggregate": {
            "status": "completed",
            "count": len(routed),
            "results": routed,
        }
    }


def build_graph():
    builder = StateGraph(SoulGraphState)
    builder.add_node("decompose", decompose)
    builder.add_node("route", route)
    builder.add_node("aggregate", aggregate)
    builder.add_edge(START, "decompose")
    builder.add_edge("decompose", "route")
    builder.add_edge("route", "aggregate")
    builder.add_edge("aggregate", END)
    return builder.compile()


GRAPH = build_graph()
