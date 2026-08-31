import os
from typing import Any

from google import genai

from .base import BaseProvider


class GeminiProvider(BaseProvider):
    """Google Gemini provider for SuperAGI generation."""

    DEFAULT_MODEL = "gemini-2.5-flash"

    def __init__(self, model: str | None = None) -> None:
        api_key = (os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY") or "").strip()
        if not api_key:
            raise RuntimeError("GOOGLE_API_KEY or GEMINI_API_KEY is not configured")

        self.client = genai.Client(api_key=api_key)
        self.model = model or os.getenv("GEMINI_MODEL", self.DEFAULT_MODEL)

    def generate(self, prompt: str, **kwargs: Any) -> str:
        response = self.client.models.generate_content(
            model=self.model,
            contents=prompt,
            **kwargs,
        )
        text = getattr(response, "text", None)
        if not text:
            raise RuntimeError("Gemini returned an empty response")
        return text.strip()
