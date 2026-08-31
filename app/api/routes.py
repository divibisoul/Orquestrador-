from __future__ import annotations

from typing import Any, Dict

from ..core.orchestrator import Orchestrator


orchestrator = Orchestrator()


def health() -> Dict[str, str]:
    return {"status": "ok"}


def tasks() -> Dict[str, Any]:
    return {"tasks": orchestrator.list_tasks()}


def execute(name: str, payload: Dict[str, Any] | None = None) -> Dict[str, Any]:
    result = orchestrator.execute(name, payload)
    return {"task": name, "result": result}
