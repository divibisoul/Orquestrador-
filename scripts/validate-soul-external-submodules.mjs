import { readFile } from "node:fs/promises";
import { execFileSync } from "node:child_process";

const run = (args) => execFileSync("git", args, { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
const registry = JSON.parse(await readFile("integrations/external-capabilities.json", "utf8"));
const gitmodules = await readFile(".gitmodules", "utf8");
const configured = new Map();
for (const block of gitmodules.split(/^\[submodule "/m).slice(1)) {
  const name = block.split("]")[0].replace(/"$/, "");
  const path = (block.match(/^\s*path\s*=\s*(.+)$/m) || [])[1]?.trim();
  const url = (block.match(/^\s*url\s*=\s*(.+)$/m) || [])[1]?.trim();
  if (name && path) configured.set(path, { name, url });
}

const errors = [];
const primary = registry.repositories ?? [];
const complementary = registry.complementary_repositories ?? [];
const activeSources = [...primary, ...complementary];
const activeIDs = new Set(activeSources.map(x => x.id));
const projectedCandidates = [...(registry.expansion_repositories ?? []), ...(registry.grce_complements ?? [])];
const projectedByID = new Map();
for (const source of projectedCandidates) {
  if (activeIDs.has(source.id)) {
    const active = activeSources.find(x => x.id === source.id);
    if (active && (active.revision !== source.revision || active.source !== source.source)) {
      errors.push(`ACTIVE_EXPANSION_SOURCE_CONFLICT:${source.id}`);
    }
    continue;
  }
  const prior = projectedByID.get(source.id);
  if (prior && (prior.revision !== source.revision || prior.source !== source.source)) {
    errors.push(`PROJECTED_SOURCE_CONFLICT:${source.id}`);
    continue;
  }
  projectedByID.set(source.id, source);
}
const projected = [...projectedByID.values()];
const sources = [...activeSources, ...projected];
const expectedPaths = new Set(sources.map(provider => `integrations/external/${provider.id}`));
const configuredExternalPaths = [...configured.keys()].filter(path => path.startsWith("integrations/external/"));
if (primary.length !== 16) errors.push(`PRIMARY_REGISTRY_COUNT:${primary.length}`);
if (complementary.length !== 6) errors.push(`COMPLEMENTARY_REGISTRY_COUNT:${complementary.length}`);
if (projected.length !== 9) errors.push(`PROJECTED_REPOSITORY_COUNT:${projected.length}`);
if (sources.length !== 31) errors.push(`TOTAL_EXTERNAL_REPOSITORY_COUNT:${sources.length}`);
for (const path of configuredExternalPaths) {
  if (!expectedPaths.has(path)) errors.push(`GITMODULE_NOT_IN_CANONICAL_REGISTRY:${path}`);
}
for (const path of expectedPaths) {
  if (!configured.has(path)) errors.push(`CANONICAL_REPOSITORY_WITHOUT_GITMODULE:${path}`);
}

const normalizeURL = (value) => String(value ?? "").replace(/\.git$/, "").replace(/\/$/, "");
for (const provider of sources) {
  const id = provider.id;
  const path = `integrations/external/${id}`;
  const cfg = configured.get(path);
  if (!cfg) {
    errors.push(`GITMODULE_MISSING:${id}`);
  } else if (normalizeURL(cfg.url) !== normalizeURL(provider.source)) {
    errors.push(`GITMODULE_SOURCE_MISMATCH:${id}:expected=${provider.source}:actual=${cfg.url || "MISSING"}`);
  }
  const treeLine = run(["ls-tree", "HEAD", path]);
  const parts = treeLine.split(/\s+/);
  if (parts[0] !== "160000" || parts[1] !== "commit" || parts[2] !== provider.revision) {
    errors.push(`GITLINK_MISMATCH:${id}:expected=${provider.revision}:tree=${treeLine || "MISSING"}`);
  }
  let head = "";
  try { head = run(["-C", path, "rev-parse", "HEAD"]); }
  catch { errors.push(`SUBMODULE_NOT_MATERIALIZED:${id}`); continue; }
  if (head !== provider.revision) errors.push(`SUBMODULE_HEAD_MISMATCH:${id}:expected=${provider.revision}:actual=${head}`);
  let status = "";
  try { status = run(["submodule", "status", "--", path]); } catch {}
  if (status.startsWith("-") || status.startsWith("+") || status.startsWith("U")) {
    errors.push(`SUBMODULE_STATUS_NOT_EXACT:${id}:${status}`);
  }
  try { run(["-C", path, "diff", "--quiet"]); }
  catch { errors.push(`SUBMODULE_WORKTREE_DIRTY:${id}`); }
  try { run(["-C", path, "diff", "--cached", "--quiet"]); }
  catch { errors.push(`SUBMODULE_INDEX_DIRTY:${id}`); }
}

if (errors.length) {
  console.error(JSON.stringify({ ok:false, errors }, null, 2));
  process.exit(2);
}
console.log(JSON.stringify({
  ok:true,
  primary_upstreams:primary.length,
  complementary_sources:complementary.length,
  projected_expansions:projected.length,
  total_pinned_external_repositories:sources.length,
  state:"MATERIALIZED_AND_EXACT",
  rule:"all external submodules must be registered; gitlink SHA == registry revision == checked-out submodule HEAD"
}, null, 2));
