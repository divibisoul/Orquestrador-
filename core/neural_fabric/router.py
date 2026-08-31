from __future__ import annotations

from dataclasses import dataclass
from typing import Any

from core.superagi.generator.providers import BaseProvider, GeminiProvider


@dataclass(frozen=True)
class ProviderEntry:
    provider: BaseProvider
    priority: int


class ProviderRouter:
    """Experimental Python provider router; isolated from Go runtime contracts."""

    def __init__(self, providers: list[ProviderEntry] | None = None) -> None:
        self.providers = sorted(
            providers or [ProviderEntry(GeminiProvider(), priority=1)],
            key=lambda item: item.priority,
        )

    def generate(self, prompt: str, **kwargs: Any) -> str:
        last_error: Exception | None = None
        for entry in self.providers:
            try:
                return entry.provider.generate(prompt, **kwargs)
            except Exception as exc:  # provider fallback boundary
                last_error = exc
        if last_error is not None:
            raise last_error
        raise RuntimeError("No generation providers are configured")
