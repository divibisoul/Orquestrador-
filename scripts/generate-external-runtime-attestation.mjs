import { execFileSync } from "node:child_process";
import { realpathSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";

const repositoryRoot = process.cwd();
const sourcePrefix = "integrations/external/";
const attestationPath = path.join(repositoryRoot, "config/soul-external-runtime-attestation.json");

function git(args, cwd = repositoryRoot) {
  return execFileSync("git", args, { cwd, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
}

function fail(message) {
  process.stderr.write(`EXTERNAL_RUNTIME_ATTESTATION_ERROR: ${message}\n`);
  process.exit(1);
}

const registry = JSON.parse(readFileSync(path.join(repositoryRoot, "integrations/external-capabilities.json"), "utf8"));
const tiers = ["repositories", "complementary_repositories", "expansion_repositories", "grce_complements"];
const sources = tiers.flatMap((tier) => registry[tier] ?? []);
const sourceIDs = sources.map((item) => item.id);
if (sources.length !== 32 || new Set(sourceIDs).size !== sources.length) {
  fail(`registered source catalog mismatch: count=${sources.length}`);
}

const treeLines = git(["ls-tree", "-r", "HEAD", "integrations/external"]).split("\n").filter(Boolean);
const treeEntries = treeLines.map((line) => {
  const tab = line.indexOf("\t");
  if (tab < 0) fail(`malformed git tree entry: ${line}`);
  const [mode, type, sha] = line.slice(0, tab).split(/\s+/);
  return { mode, type, sha, path: line.slice(tab + 1) };
});
const rootGitlinks = treeEntries.filter((entry) => {
  if (entry.mode !== "160000" || entry.type !== "commit" || !entry.path.startsWith(sourcePrefix)) return false;
  const id = entry.path.slice(sourcePrefix.length);
  return !id.includes("/") && id !== "recurs";
});
const expectedPaths = new Set(sourceIDs.map((id) => `${sourcePrefix}${id}`));
const actualPaths = new Set(rootGitlinks.map((entry) => entry.path));
if (actualPaths.size !== expectedPaths.size || [...expectedPaths].some((p) => !actualPaths.has(p))) {
  fail(`root gitlinks disagree with registry: expected=${expectedPaths.size} actual=${actualPaths.size}`);
}

const repositories = {};
for (const source of sources) {
  const id = source.id;
  const expected = source.revision;
  if (typeof expected !== "string" || !/^[0-9a-f]{40}$/i.test(expected)) {
    fail(`invalid registered revision: ${id}`);
  }
  const relativeRoot = `${sourcePrefix}${id}`;
  const absoluteRoot = path.join(repositoryRoot, relativeRoot);
  let stat;
  try { stat = statSync(absoluteRoot); } catch { fail(`source worktree missing: ${id}`); }
  if (!stat.isDirectory()) fail(`source worktree is not a directory: ${id}`);
  const actualTop = realpathSync(git(["-C", absoluteRoot, "rev-parse", "--show-toplevel"]));
  if (actualTop !== realpathSync(absoluteRoot)) fail(`git escaped provider worktree: ${id}`);
  const actualHead = git(["-C", absoluteRoot, "rev-parse", "HEAD"]);
  const gitlink = rootGitlinks.find((entry) => entry.path === relativeRoot);
  if (!gitlink || gitlink.sha !== expected || actualHead !== expected) {
    fail(`source pin mismatch: ${id}: registry=${expected} gitlink=${gitlink?.sha ?? "missing"} worktree=${actualHead}`);
  }
  repositories[id] = expected;
}

const attestation = {
  schema_version: "soul.external.runtime-attestation.v1",
  generated_from_materialized_git: true,
  repositories,
};
writeFileSync(attestationPath, `${JSON.stringify(attestation, null, 2)}\n`, { encoding: "utf8", mode: 0o644 });
process.stdout.write(`EXTERNAL_RUNTIME_ATTESTATION: PASS (${Object.keys(repositories).length} exact gitlink/worktree pins)\n`);
process.stdout.write(`EXTERNAL_RUNTIME_ATTESTATION_SHA256_INPUT: ${Object.keys(repositories).sort().join(",")}\n`);
