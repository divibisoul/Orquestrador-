from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

from core.superagi.generator.providers import GeminiProvider


class GeminiAPIHandler(BaseHTTPRequestHandler):
    provider = GeminiProvider()

    def _send_json(self, status: int, payload: dict[str, Any]) -> None:
        data = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self) -> None:  # noqa: N802
        if self.path == "/health":
            self._send_json(200, {"ok": True, "provider": "gemini"})
            return
        self._send_json(404, {"ok": False, "error": "NOT_FOUND"})

    def do_POST(self) -> None:  # noqa: N802
        if self.path != "/v1/generate":
            self._send_json(404, {"ok": False, "error": "NOT_FOUND"})
            return

        try:
            length = int(self.headers.get("Content-Length", "0"))
            raw = self.rfile.read(length)
            body = json.loads(raw or b"{}")
            prompt = body.get("prompt")
            if not isinstance(prompt, str) or not prompt.strip():
                self._send_json(400, {"ok": False, "error": "PROMPT_REQUIRED"})
                return

            text = self.provider.generate(prompt, **(body.get("generation_config") or {}))
            self._send_json(200, {"ok": True, "provider": "gemini", "model": self.provider.model, "text": text})
        except Exception as exc:  # noqa: BLE001 - API boundary
            self._send_json(502, {"ok": False, "provider": "gemini", "error": type(exc).__name__, "detail": str(exc)})

    def log_message(self, format: str, *args: object) -> None:
        return


def serve() -> None:
    host = os.getenv("ORCHESTRATOR_API_HOST", "0.0.0.0")
    port = int(os.getenv("ORCHESTRATOR_API_PORT", "8090"))
    ThreadingHTTPServer((host, port), GeminiAPIHandler).serve_forever()


if __name__ == "__main__":
    serve()
