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
if (registry.repositories.length !== 16) errors.push(`REGISTRY_COUNT:${registry.repositories.length}`);
for (const provider of registry.repositories) {
  const id = provider.id;
  const path = `integrations/external/${id}`;
  const cfg = configured.get(path);
  if (!cfg) errors.push(`GITMODULE_MISSING:${id}`);
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
console.log(JSON.stringify({ ok:true, upstreams:16, state:"MATERIALIZED_AND_EXACT", rule:"gitlink SHA == registry revision == checked-out submodule HEAD" }, null, 2));
