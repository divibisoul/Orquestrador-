from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Callable, Dict, List


@dataclass
class Task:
    name: str
    handler: Callable[[Dict[str, Any]], Any]
    metadata: Dict[str, Any] = field(default_factory=dict)


class Orchestrator:
    """Small Python orchestration facade for API-facing integrations."""

    def __init__(self) -> None:
        self._tasks: Dict[str, Task] = {}

    def register(self, task: Task) -> None:
        if not task.name.strip():
            raise ValueError("task name is required")
        if task.name in self._tasks:
            raise ValueError("task already registered")
        self._tasks[task.name] = task

    def execute(self, name: str, payload: Dict[str, Any] | None = None) -> Any:
        task = self._tasks.get(name)
        if task is None:
            raise KeyError(name)
        return task.handler(dict(payload or {}))

    def list_tasks(self) -> List[str]:
        return sorted(self._tasks)
