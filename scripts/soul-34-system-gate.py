#!/usr/bin/env python3
from __future__ import annotations

import json
import os
import re
import sys
import urllib.error
import urllib.request
from urllib.parse import quote

ROOT = "https://raw.githubusercontent.com"
API = "https://api.github.com"
EXPECTED_NATIVE = 9
EXPECTED_UPSTREAM = 25
SHA_RE = re.compile(r"^[0-9a-f]{40}$")

def fail(message: str) -> None:
    print(f"SOUL-34 GATE FAIL: {message}", file=sys.stderr)
    raise SystemExit(1)

def get(url: str) -> bytes:
    headers = {"Accept": "application/json"}
    token = os.environ.get("GITHUB_TOKEN", "").strip()
    if token:
        headers["Authorization"] = f"Bearer {token}"
    request = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            return response.read()
    except urllib.error.HTTPError as exc:
        fail(f"HTTP {exc.code}: {url}")
    except urllib.error.URLError as exc:
        fail(f"network error for {url}: {exc}")

with open("integrations/soul-34-system-integration.json", "r", encoding="utf-8") as handle:
    system = json.load(handle)

native = system.get("native_components", [])
upstreams = system.get("upstream_sources", [])

if system.get("repository_count") != 34:
    fail(f"repository_count={system.get('repository_count')} expected 34")
if len(native) != EXPECTED_NATIVE:
    fail(f"native components={len(native)} expected {EXPECTED_NATIVE}")
if system.get("upstream_count") != EXPECTED_UPSTREAM:
    fail(f"upstream_count={system.get('upstream_count')} expected {EXPECTED_UPSTREAM}")
if len(upstreams) != EXPECTED_UPSTREAM:
    fail(f"upstream records={len(upstreams)} expected {EXPECTED_UPSTREAM}")

upstream_ids = [str(x.get("id", "")).strip() for x in upstreams]
if len(set(upstream_ids)) != EXPECTED_UPSTREAM:
    fail("duplicate upstream ids in canonical system contract")

verified_upstreams = 0
for source in upstreams:
    repo_name = str(source.get("repository", "")).strip()
    revision = str(source.get("revision", "")).strip()
    if "/" not in repo_name:
        fail(f"invalid upstream repository: {repo_name}")
    if not SHA_RE.fullmatch(revision):
        fail(f"invalid upstream revision for {source.get('id')}: {revision}")
    payload = get(f"{API}/repos/{repo_name}/git/commits/{revision}")
    data = json.loads(payload.decode("utf-8"))
    if str(data.get("sha", "")).lower() != revision.lower():
        fail(f"upstream revision mismatch for {source.get('id')}")
    verified_upstreams += 1

verified_components = 0
for component in native:
    cid = str(component.get("id", "")).strip()
    repo_name = str(component.get("repository", "")).strip()
    branch = str(component.get("branch", "")).strip()
    fabric_path = str(component.get("fabric", "")).strip()
    resident_path = str(component.get("resident_agent_path", "")).strip()
    if not all((cid, repo_name, branch, fabric_path, resident_path)):
        fail(f"incomplete component contract for {cid}")
    manifest_url = f"{ROOT}/{repo_name}/{quote(branch, safe='/')}/integrations/soul-25-augmentation.json"
    fabric_url = f"{ROOT}/{repo_name}/{quote(branch, safe='/')}/{quote(fabric_path, safe='/')}"
    resident_url = f"{ROOT}/{repo_name}/{quote(branch, safe='/')}/{quote(resident_path, safe='/')}"
    try:
        manifest_bytes = get(manifest_url)
        fabric_bytes = get(fabric_url)
        resident_bytes = get(resident_url)
    except SystemExit:
        raise
    manifest = json.loads(manifest_bytes.decode("utf-8"))
    if manifest.get("component_id") != cid:
        fail(f"{cid}: manifest component_id={manifest.get('component_id')}")
    if manifest.get("external_provider_count") != EXPECTED_UPSTREAM:
        fail(f"{cid}: manifest provider count={manifest.get('external_provider_count')}")
    providers = manifest.get("external_providers", [])
    if len(providers) != EXPECTED_UPSTREAM:
        fail(f"{cid}: manifest external_providers={len(providers)}")
    ids = {str(x.get("id", "")).strip() for x in providers}
    if ids != set(upstream_ids):
        missing = sorted(set(upstream_ids) - ids)
        extra = sorted(ids - set(upstream_ids))
        fail(f"{cid}: provider set mismatch missing={missing} extra={extra}")
    if manifest.get("native_authority") != "preserved":
        fail(f"{cid}: native_authority not preserved")
    if len(fabric_bytes) == 0:
        fail(f"{cid}: empty fabric source")
    if len(resident_bytes) == 0:
        fail(f"{cid}: empty resident agent source")
    verified_components += 1

print(f"SOUL-34 GATE PASS: {verified_components}/{EXPECTED_NATIVE} native components, {verified_upstreams}/{EXPECTED_UPSTREAM} upstream revisions, 34 repositories.")
