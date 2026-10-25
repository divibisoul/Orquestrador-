import json
import os
import unittest
from unittest.mock import patch

from graph import _call_n07, aggregate, build_graph, decompose


class FakeResponse:
    def __init__(self, payload, status=200):
        self.payload = payload
        self.status = status

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, traceback):
        return False

    def read(self, limit):
        return json.dumps(self.payload).encode("utf-8")[:limit]


class GraphTests(unittest.TestCase):
    def test_graph_compiles(self):
        self.assertIsNotNone(build_graph())

    def test_decompose_normalizes_tasks_and_preserves_correlation(self):
        result = decompose({
            "correlationId": "mesh-correlation-001",
            "tasks": [{"id": "a", "prompt": "hello"}],
        })
        self.assertEqual(
            result["tasks"],
            [{
                "id": "a",
                "prompt": "hello",
                "model": "soul-auto",
                "correlationId": "mesh-correlation-001",
            }],
        )
        self.assertEqual(result["correlationId"], "mesh-correlation-001")
        self.assertEqual(result["routed"], [])

    def test_decompose_rejects_noncanonical_model(self):
        with self.assertRaisesRegex(ValueError, "only accepts model soul-auto"):
            decompose({"tasks": [{"id": "a", "prompt": "hello", "model": "gpt-5"}]})

    def test_call_n07_sends_and_verifies_real_returned_correlation(self):
        correlation = "mesh-correlation-real-002"
        request_seen = {}

        def fake_urlopen(request, timeout):
            request_seen["url"] = request.full_url
            request_seen["auth"] = request.get_header("Authorization")
            request_seen["request_id"] = request.get_header("X-request-id")
            request_seen["timeout"] = timeout
            payload = {
                "choices": [{"message": {"content": "real test response"}}],
                "metadata": {"n07_owner": "N02", "correlation_id": correlation},
            }
            return FakeResponse(payload)

        with patch.dict(os.environ, {"SOUL_N07_URL": "http://n07.test", "N07_APP_TOKEN": "test-token"}):
            with patch("urllib.request.urlopen", side_effect=fake_urlopen):
                result = _call_n07({
                    "id": "task-1",
                    "prompt": "hello",
                    "model": "soul-auto",
                    "correlationId": correlation,
                })

        self.assertEqual(request_seen["url"], "http://n07.test/v1/chat/completions")
        self.assertEqual(request_seen["auth"], "Bearer test-token")
        self.assertEqual(request_seen["request_id"], correlation)
        self.assertEqual(result["text"], "real test response")
        self.assertEqual(result["metadata"]["n07_owner"], "N02")
        self.assertEqual(result["correlationId"], correlation)

    def test_call_n07_rejects_wrong_returned_correlation(self):
        with patch.dict(os.environ, {"SOUL_N07_URL": "http://n07.test", "N07_APP_TOKEN": "test-token"}):
            with patch("urllib.request.urlopen", return_value=FakeResponse({
                "choices": [{"message": {"content": "wrong correlation"}}],
                "metadata": {"correlation_id": "different-correlation"},
            })):
                with self.assertRaisesRegex(RuntimeError, "correlation ID mismatch"):
                    _call_n07({
                        "id": "task-2",
                        "prompt": "hello",
                        "model": "soul-auto",
                        "correlationId": "expected-correlation",
                    })

    def test_aggregate_preserves_routed_results(self):
        routed = [{
            "id": "a",
            "status": "routed",
            "text": "ok",
            "metadata": {},
            "correlationId": "mesh-correlation-003",
        }]
        result = aggregate({"tasks": [], "routed": routed})
        self.assertEqual(result["aggregate"]["count"], 1)
        self.assertEqual(result["aggregate"]["results"], routed)


if __name__ == "__main__":
    unittest.main()
