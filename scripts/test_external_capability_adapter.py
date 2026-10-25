#!/usr/bin/env python3
"""Contract regression for the N07 external adapter's LangGraph execution gate."""
import json
import os
from pathlib import Path
import subprocess
import sys
import unittest


class ExternalCapabilityAdapterTests(unittest.TestCase):
    def test_langgraph_never_reports_pass_without_n07_authentication(self):
        repository_root = Path(__file__).resolve().parents[1]
        adapter = Path(__file__).with_name("external-capability-adapter.py")
        request = {
            "provider": "langgraph",
            "mode": "execute",
            "operation": "workflow.invoke",
            "root": str(repository_root),
            "correlationId": "external-adapter-auth-regression",
            "prompt": "A test request must be blocked before external execution.",
        }
        env = os.environ.copy()
        env.pop("N07_APP_TOKEN", None)

        completed = subprocess.run(
            [sys.executable, str(adapter)],
            input=json.dumps(request),
            text=True,
            capture_output=True,
            timeout=10,
            check=False,
            env=env,
            cwd=repository_root,
        )

        self.assertNotEqual(completed.returncode, 0, completed.stdout)
        result = json.loads(completed.stdout)
        self.assertEqual(result.get("state"), "BLOCKED")
        self.assertEqual(result.get("code"), "LANGGRAPH_N07_APP_TOKEN_REQUIRED")
        self.assertNotIn('"state": "PASS"', completed.stdout)


if __name__ == "__main__":
    unittest.main()
