import { readFile } from "node:fs/promises";
import { execFileSync } from "node:child_process";

const run = (args, cwd = process.cwd()) =>
  execFileSync("git", args, { cwd, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();

const gitmodules = await readFile(".gitmodules", "utf8");
const paths = [];
for (const block of gitmodules.split(/^\[submodule "/m).slice(1)) {
  const path = (block.match(/^\s*path\s*=\s*(.+)$/m) || [])[1]?.trim();
  if (path) paths.push(path);
}

if (paths.length === 0) {
  console.log("TOP_LEVEL_SUBMODULES: NONE");
  process.exit(0);
}

for (const path of paths) {
  console.log("MATERIALIZE_TOP_LEVEL:", path);
  run(["submodule", "update", "--init", "--", path]);
}

const autogenesis = "integrations/external/autogenesis";
try {
  run(["rev-parse", "--git-dir"], autogenesis);
  const nested = run(["ls-tree", "HEAD", "datasets/hle"], autogenesis);
  const expected = "160000 commit 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29";
  if (!nested.includes(expected)) {
    throw new Error(`NESTED_SUBMODULE_PIN_MISMATCH:${nested}`);
  }
  // GitHub Actions checkout cleanup may inspect nested metadata recursively.
  // Materialize the upstream's own missing .gitmodules only in the ephemeral CI
  // worktree; the upstream repository and its top-level registration remain unchanged.
  const nestedGitmodules = "[submodule \"datasets/hle\"]\n  path = datasets/hle\n  url = https://huggingface.co/datasets/cais/hle\n";
  const { writeFile } = await import("node:fs/promises");
  await writeFile(`${autogenesis}/.gitmodules`, nestedGitmodules, "utf8");
  run(["config", "submodule.datasets/hle.url", "https://huggingface.co/datasets/cais/hle"], autogenesis);
  console.log("NESTED_SUBMODULE_METADATA: PROJECTED/BLOCKED");
  console.log("NESTED_SUBMODULE_REASON: Autogenesis pins HLE but its own .gitmodules is absent.");
  console.log("NESTED_SUBMODULE_PROVENANCE: HLE revision 5a81a4c7271a2a2a312b9a690f0c2fde837e4c29");
  console.log("NESTED_SUBMODULE_ACTION: CI-local URL repair only; upstream remains unchanged.");
} catch (error) {
  if (error?.code === "ENOENT" || String(error?.message || "").includes("not found")) {
    // Autogenesis is optional on branches that do not declare it.
  } else if (error?.message) {
    throw error;
  }
}

console.log("TOP_LEVEL_SUBMODULES: MATERIALIZED");
