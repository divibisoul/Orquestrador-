from __future__ import annotations

from dataclasses import dataclass
from typing import Any

from core.superagi.generator.providers import BaseProvider, GeminiProvider


@dataclass(frozen=True)
class ProviderEntry:
    provider: BaseProvider
    priority: int


class ProviderRouter:
    """Python provider router with Gemini as the highest-priority provider.

    This adapter is intentionally isolated from the existing Go Neural Fabric.
    It provides the requested Python 3.10+ provider-routing surface without
    changing the existing Go runtime contracts.
    """

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
            except Exception as exc:  # noqa: BLE001 - provider fallback boundary
                last_error = exc

        if last_error is not None:
            raise last_error
        raise RuntimeError("No generation providers are configured")


# Gemini is the default first-priority provider.
providers = [ProviderEntry(GeminiProvider(), priority=1)]
router = ProviderRouter(providers)
