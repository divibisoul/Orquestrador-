import { readFile } from "node:fs/promises";
import { realpathSync } from "node:fs";
import { execFileSync } from "node:child_process";

const run = (args) => execFileSync("git", args, { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
const isExactWorktree = (root) => {
  try {
    const actual = realpathSync(run(["-C", root, "rev-parse", "--show-toplevel"]));
    return actual === realpathSync(root);
  } catch {
    // Git can walk upward from an unmaterialized submodule and report N07.
    return false;
  }
};
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
const sources = [...primary, ...complementary];
const historicalComplementaryIds = [
  "autogenesis", "octos", "hora-graph-core", "mycelium", "prime-agent", "cuda-oxide"
].sort();
const expandedComplementaryIds = [
  ...historicalComplementaryIds,
  "bijux-dag-runtime", "ouro-loop", "recuris", "fedml", "hivemind",
  "temporal", "cognitive-workspace", "ravana", "ray", "nats-go"
].sort();
const complementaryIds = complementary.map(x => x.id).sort();
if (primary.length !== 16) errors.push(`PRIMARY_REGISTRY_COUNT:${primary.length}`);
if (complementary.length !== 6 && complementary.length !== 16) {
  errors.push(`COMPLEMENTARY_REGISTRY_COUNT:${complementary.length}`);
} else {
  const expected = complementary.length === 6 ? historicalComplementaryIds : expandedComplementaryIds;
  if (JSON.stringify(complementaryIds) !== JSON.stringify(expected)) errors.push("COMPLEMENTARY_REGISTRY_SET_MISMATCH");
}
if (sources.length !== primary.length + complementary.length) errors.push("TOTAL_REGISTRY_COUNT_MISMATCH");
if (new Set(sources.map(x => x.id)).size !== sources.length) errors.push("DUPLICATE_SOURCE_IDS");
for (const provider of sources) {
  const id = provider.id;
  const path = `integrations/external/${id}`;
  const cfg = configured.get(path);
  if (!cfg) {
    errors.push(`GITMODULE_MISSING:${id}`);
  } else {
    const normalizeURL = (value) => value.replace(/\.git$/, "").replace(/\/$/, "");
    if (normalizeURL(cfg.url || "") !== normalizeURL(provider.source || "")) {
      errors.push(`GITMODULE_SOURCE_MISMATCH:${id}:expected=${provider.source}:actual=${cfg.url || "MISSING"}`);
    }
  }
  const treeLine = run(["ls-tree", "HEAD", path]);
  const parts = treeLine.split(/\s+/);
  if (parts[0] !== "160000" || parts[1] !== "commit" || parts[2] !== provider.revision) {
    errors.push(`GITLINK_MISMATCH:${id}:expected=${provider.revision}:tree=${treeLine || "MISSING"}`);
  }
  // Prove -C PATH resolves to the submodule itself before reading HEAD.
  // Otherwise Git may return the parent N07 SHA for an empty path.
  if (!isExactWorktree(path)) {
    errors.push(`SUBMODULE_NOT_MATERIALIZED:${id}`);
    continue;
  }
  const head = run(["-C", path, "rev-parse", "HEAD"]);
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
  total_sources:sources.length,
  state:"MATERIALIZED_AND_EXACT",
  rule:"gitlink SHA == registry revision == checked-out submodule HEAD"
}, null, 2));
