import os
from typing import Any

from google import genai

from .base import BaseProvider


class GeminiProvider(BaseProvider):
    """Gemini provider with lazy credential initialization."""

    DEFAULT_MODEL = "gemini-2.5-flash"

    def __init__(self, model: str | None = None) -> None:
        self.model = model or os.getenv("GEMINI_MODEL", self.DEFAULT_MODEL)
        self._client: genai.Client | None = None

    def _client_or_create(self) -> genai.Client:
        if self._client is not None:
            return self._client
        api_key = (os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY") or "").strip()
        if not api_key:
            raise RuntimeError("Gemini provider is not configured")
        self._client = genai.Client(api_key=api_key)
        return self._client

    def generate(self, prompt: str, **kwargs: Any) -> str:
        if not isinstance(prompt, str) or not prompt.strip():
            raise ValueError("prompt is required")
        response = self._client_or_create().models.generate_content(
            model=self.model,
            contents=prompt,
            **kwargs,
        )
        text = getattr(response, "text", None)
        if not text:
            raise RuntimeError("Gemini returned no text")
        return text.strip()
