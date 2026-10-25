import { readFile, writeFile } from "node:fs/promises";
import { execFileSync } from "node:child_process";

const run = (args, cwd = process.cwd()) =>
  execFileSync("git", args, { cwd, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();

const gitmodules = await readFile(".gitmodules", "utf8");
const registry = JSON.parse(await readFile("integrations/external-capabilities.json", "utf8"));
const tiers = ["repositories", "complementary_repositories", "expansion_repositories", "grce_complements"];
const canonicalIDs = new Set(tiers.flatMap((tier) => (registry[tier] ?? []).map((entry) => entry.id)));
const canonicalPaths = new Set([...canonicalIDs].map((id) => `integrations/external/${id}`));
const paths = [];
const preservedAliases = [];

for (const block of gitmodules.split(/^\[submodule "/m).slice(1)) {
  const path = (block.match(/^\s*path\s*=\s*(.+)$/m) || [])[1]?.trim();
  if (!path) continue;
  if (path.startsWith("integrations/external/") && !canonicalPaths.has(path)) {
    // Legacy link remains in Git for provenance but is not a canonical runtime source.
    preservedAliases.push(path);
    continue;
  }
  paths.push(path);
}

if (paths.length === 0) {
  console.log("TOP_LEVEL_SUBMODULES: NONE");
  process.exit(0);
}
const missing = [...canonicalPaths].filter((path) => !paths.includes(path));
if (missing.length) {
  throw new Error(`CANONICAL_SUBMODULE_GITLINKS_MISSING:${missing.join(",")}`);
}

// Fetch canonical upstreams concurrently with bounded fan-out. Historical aliases
// stay versioned in Git but are not materialized as an extra source.
console.log(`MATERIALIZE_TOP_LEVEL_COUNT: ${paths.length}`);
console.log("MATERIALIZE_TOP_LEVEL_JOBS: 4");
for (const path of paths) console.log("MATERIALIZE_TOP_LEVEL:", path);
for (const path of preservedAliases) console.log("PRESERVED_LEGACY_SUBMODULE_NOT_MATERIALIZED:", path);
run(["submodule", "update", "--init", "--jobs=4", "--", ...paths]);

const autogenesis = "integrations/external/autogenesis";
try {
  run(["rev-parse", "--git-dir"], autogenesis);
  const nested = run(["ls-tree", "HEAD", "datasets/hle"], autogenesis);
  const expected = "160000 commit 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29";
  if (!nested.includes(expected)) {
    throw new Error(`NESTED_SUBMODULE_PIN_MISMATCH:${nested}`);
  }
  // Repair only ephemeral CI-local metadata; upstream source is untouched.
  const nestedGitmodules = "[submodule \"datasets/hle\"]\n  path = datasets/hle\n  url = https://huggingface.co/datasets/cais/hle\n";
  await writeFile(`${autogenesis}/.gitmodules`, nestedGitmodules, "utf8");
  run(["config", "submodule.datasets/hle.url", "https://huggingface.co/datasets/cais/hle"], autogenesis);
  console.log("NESTED_SUBMODULE_METADATA: PROJECTED/BLOCKED");
  console.log("NESTED_SUBMODULE_REASON: Autogenesis pins HLE but its own .gitmodules is absent.");
  console.log("NESTED_SUBMODULE_PROVENANCE: HLE revision 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29");
  console.log("NESTED_SUBMODULE_ACTION: CI-local URL repair only; upstream remains unchanged.");
} catch (error) {
  if (error?.code === "ENOENT" || String(error?.message || "").includes("not found")) {
    // Autogenesis optional for branches that do not declare it.
  } else if (error?.message) {
    throw error;
  }
}

console.log("TOP_LEVEL_SUBMODULES: MATERIALIZED");
