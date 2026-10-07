from __future__ import annotations

import json
import sys

from graph import GRAPH


def main() -> int:
    raw = sys.stdin.read()
    if not raw.strip():
        raise SystemExit("stdin must contain a JSON object")
    result = GRAPH.invoke(json.loads(raw))
    json.dump(result, sys.stdout, ensure_ascii=False, separators=(",", ":"))
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
