from __future__ import annotations

import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

from core.superagi.generator.providers import GeminiProvider


class GeminiAPIHandler(BaseHTTPRequestHandler):
    provider = GeminiProvider()

    def _authorized(self) -> bool:
        expected = os.getenv("ORCHESTRATOR_API_KEY", "").strip()
        supplied = self.headers.get("X-API-Key", "").strip()
        return bool(expected) and supplied == expected

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
        if not self._authorized():
            self._send_json(401, {"ok": False, "error": "UNAUTHORIZED"})
            return
        try:
            length = int(self.headers.get("Content-Length", "0"))
            if length <= 0 or length > 1_000_000:
                self._send_json(400, {"ok": False, "error": "INVALID_BODY"})
                return
            body = json.loads(self.rfile.read(length))
            prompt = body.get("prompt")
            if not isinstance(prompt, str) or not prompt.strip():
                self._send_json(400, {"ok": False, "error": "PROMPT_REQUIRED"})
                return
            config = body.get("generation_config")
            if config is not None and not isinstance(config, dict):
                self._send_json(400, {"ok": False, "error": "INVALID_GENERATION_CONFIG"})
                return
            text = self.provider.generate(prompt, **(config or {}))
            self._send_json(200, {"ok": True, "provider": "gemini", "model": self.provider.model, "text": text})
        except (ValueError, TypeError, json.JSONDecodeError):
            self._send_json(400, {"ok": False, "error": "INVALID_REQUEST"})
        except Exception:
            self._send_json(502, {"ok": False, "error": "PROVIDER_ERROR"})

    def log_message(self, format: str, *args: object) -> None:
        return


def serve() -> None:
    host = os.getenv("ORCHESTRATOR_API_HOST", "127.0.0.1")
    port = int(os.getenv("ORCHESTRATOR_API_PORT", "8090"))
    ThreadingHTTPServer((host, port), GeminiAPIHandler).serve_forever()


if __name__ == "__main__":
    serve()
