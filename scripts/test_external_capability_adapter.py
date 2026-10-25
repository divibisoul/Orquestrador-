#!/usr/bin/env python3
"""Contract regressions for the N07 external adapter's fail-closed boundaries."""
import json
import os
from pathlib import Path
import subprocess
import sys
import unittest


class ExternalCapabilityAdapterTests(unittest.TestCase):
    def setUp(self):
        self.repository_root = Path(__file__).resolve().parents[1]
        self.adapter = Path(__file__).with_name("external-capability-adapter.py")

    def invoke(self, request, *, overrides=None):
        env = os.environ.copy()
        for key in (
            "N07_APP_TOKEN",
            "BROWSER_USE_API_KEY",
            "OPENAI_API_KEY",
            "SOUL_EXTERNAL_OPENHANDS_URL",
            "SOUL_EXTERNAL_OPENHANDS_CONVERSATION_ID",
            "SOUL_EXTERNAL_OPENHANDS_SESSION_KEY",
            "SOUL_EXTERNAL_SUPERAGI_URL",
            "SOUL_EXTERNAL_SUPERAGI_AGENT_ID",
            "SOUL_EXTERNAL_SUPERAGI_API_KEY",
        ):
            env.pop(key, None)
        if overrides:
            env.update(overrides)
        completed = subprocess.run(
            [sys.executable, str(self.adapter)],
            input=json.dumps(request),
            text=True,
            capture_output=True,
            timeout=10,
            check=False,
            env=env,
            cwd=self.repository_root,
        )
        try:
            result = json.loads(completed.stdout)
        except Exception as exc:
            self.fail(f"adapter did not emit JSON: {exc}; stdout={completed.stdout!r}; stderr={completed.stderr!r}")
        return completed, result

    def test_langgraph_never_reports_pass_without_n07_authentication(self):
        completed, result = self.invoke({
            "provider": "langgraph",
            "mode": "execute",
            "operation": "workflow.invoke",
            "root": str(self.repository_root),
            "correlationId": "external-adapter-auth-regression",
            "prompt": "A test request must be blocked before external execution.",
        })
        self.assertNotEqual(completed.returncode, 0, completed.stdout)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "LANGGRAPH_N07_APP_TOKEN_REQUIRED")
        self.assertNotIn('"state": "PASS"', completed.stdout)

    def test_browser_use_never_executes_without_domain_allowlist(self):
        completed, result = self.invoke({
            "provider": "browser-use",
            "mode": "execute",
            "operation": "browser.execute",
            "root": str(self.repository_root),
            "task": "Do not navigate unless an explicit allowlist is provided.",
        })
        self.assertNotEqual(completed.returncode, 0)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "BROWSER_USE_DOMAIN_ALLOWLIST_REQUIRED")

    def test_openhands_request_cannot_choose_network_endpoint(self):
        completed, result = self.invoke({
            "provider": "openhands",
            "mode": "execute",
            "operation": "agent.execute",
            "root": str(self.repository_root),
            "base_url": "http://127.0.0.1:1",
            "conversation_id": "test-conversation",
            "prompt": "This endpoint must not be accepted from request metadata.",
        })
        self.assertNotEqual(completed.returncode, 0)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "OPENHANDS_BASE_URL_CONVERSATION_AND_PROMPT_REQUIRED")

    def test_superagi_request_cannot_choose_network_endpoint(self):
        completed, result = self.invoke({
            "provider": "superagi",
            "mode": "execute",
            "operation": "agent.execute",
            "root": str(self.repository_root),
            "base_url": "http://127.0.0.1:1",
            "agent_id": "test-agent",
            "goal": "This endpoint must not be accepted from request metadata.",
        }, overrides={
            "SOUL_EXTERNAL_SUPERAGI_AGENT_ID": "test-agent",
            "SOUL_EXTERNAL_SUPERAGI_API_KEY": "test-only-not-a-real-key",
        })
        self.assertNotEqual(completed.returncode, 0)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "SUPERAGI_BASE_URL_AGENT_ID_AND_API_KEY_REQUIRED")

    def test_letta_requires_explicit_agent_identity(self):
        completed, result = self.invoke({
            "provider": "letta-code",
            "mode": "execute",
            "operation": "memory.execute",
            "root": str(self.repository_root),
            "prompt": "Do not create a new persistent agent implicitly.",
        })
        self.assertNotEqual(completed.returncode, 0)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "LETTA_AGENT_ID_REQUIRED")


if __name__ == "__main__":
    unittest.main()
