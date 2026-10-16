import crypto from "node:crypto";
import fs from "node:fs";

const file = "integrations/soul-28-capability-fabric.json";
const preserved = "_preserved/soul-25-capability-fabric.json";
const d = JSON.parse(fs.readFileSync(file, "utf8"));
const preservedBytes = fs.readFileSync(preserved);
const preservedHash = "sha256:" + crypto.createHash("sha256").update(preservedBytes).digest("hex");

if (d.counts.total_nodes !== 28) throw new Error(`SOUL28_NODE_COUNT:${d.counts.total_nodes}`);
if (d.counts.edges !== 156) throw new Error(`SOUL28_EDGE_COUNT:${d.counts.edges}`);
if (d.nodes.length !== 28 || d.edges.length !== 156) throw new Error("SOUL28_ARRAY_COUNTS_MISMATCH");

const ids = new Set(d.nodes.map((x) => x.id));
for (const e of d.edges) {
  if (!ids.has(e.from) || !ids.has(e.to)) {
    throw new Error(`SOUL28_EDGE_ENDPOINT_MISSING:${e.from}->${e.to}`);
  }
}

const edgeKeys = d.edges.map((e) => `${e.from}->${e.to}`);
if (new Set(edgeKeys).size !== d.edges.length) throw new Error("SOUL28_DUPLICATE_EDGE");

const complements = ["bijux-dag-runtime", "ouro-loop", "recurs"];
for (const id of complements) {
  if (!ids.has(id)) throw new Error(`SOUL28_COMPLEMENT_MISSING:${id}`);
}

const expectedComplementEdges = [
  ["N07", "bijux-dag-runtime", "direct-affinity"],
  ["bijux-dag-runtime", "N07", "return-contract"],
  ["SARA", "bijux-dag-runtime", "federated-availability"],
  ["bijux-dag-runtime", "N01", "state-evidence"],
  ["N07", "ouro-loop", "direct-affinity"],
  ["ouro-loop", "N07", "return-contract"],
  ["SARA", "ouro-loop", "federated-availability"],
  ["ouro-loop", "N06", "agent-remediation"],
  ["N07", "recurs", "direct-affinity"],
  ["recurs", "N07", "return-contract"],
  ["N06", "recurs", "learning-support"],
  ["recurs", "N01", "state-evidence"],
];

const actualComplementEdges = new Map(
  d.edges
    .filter((e) => complements.includes(e.from) || complements.includes(e.to))
    .map((e) => [`${e.from}->${e.to}`, e])
);

if (actualComplementEdges.size !== expectedComplementEdges.length) {
  throw new Error(`SOUL28_COMPLEMENT_EDGE_COUNT:${actualComplementEdges.size}`);
}

for (const [from, to, mode] of expectedComplementEdges) {
  const edge = actualComplementEdges.get(`${from}->${to}`);
  if (!edge) throw new Error(`SOUL28_COMPLEMENT_EDGE_MISSING:${from}->${to}`);
  if (edge.mode !== mode) throw new Error(`SOUL28_COMPLEMENT_EDGE_MODE:${from}->${to}:${edge.mode}`);
  if (edge.state !== "PROJECTED") throw new Error(`SOUL28_COMPLEMENT_EDGE_STATE:${from}->${to}:${edge.state}`);
}

for (const id of complements) {
  const incident = d.edges.filter((e) => e.from === id || e.to === id).length;
  if (incident !== 4) throw new Error(`SOUL28_EDGE_DEGREE:${id}:${incident}`);
}

if (!/^sha256:[0-9a-f]{64}$/.test(d.parent_hash)) throw new Error("SOUL28_PARENT_HASH_INVALID");
if (d.parent_hash !== preservedHash) {
  throw new Error(`SOUL28_PARENT_HASH_MISMATCH:expected=${preservedHash}:actual=${d.parent_hash}`);
}

console.log(JSON.stringify({
  ok: true,
  nodes: d.nodes.length,
  edges: d.edges.length,
  parent_hash: d.parent_hash,
  preserved_parent_hash: preservedHash,
  complements,
  complement_edges: expectedComplementEdges.length,
  edge_degree: Object.fromEntries(complements.map((id) => [
    id,
    d.edges.filter((e) => e.from === id || e.to === id).length,
  ])),
}, null, 2));
