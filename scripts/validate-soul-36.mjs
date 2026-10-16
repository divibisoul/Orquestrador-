import { readFile } from "node:fs/promises";
import { createHash } from "node:crypto";

const soul25 = await readFile("integrations/soul-25-capability-fabric.json");
const preserved = await readFile("_preserved/soul-25-capability-fabric.json");
if (!soul25.equals && Buffer.compare(Buffer.from(soul25), Buffer.from(preserved)) !== 0) {
  throw new Error("SOUL25_PRESERVED_COPY_NOT_BYTE_IDENTICAL");
}
if (Buffer.compare(Buffer.from(soul25), Buffer.from(preserved)) !== 0) {
  throw new Error("SOUL25_PRESERVED_COPY_NOT_BYTE_IDENTICAL");
}

const soul36 = JSON.parse(await readFile("integrations/soul-36-capability-fabric.json", "utf8"));
const parentHash = "sha256:" + createHash("sha256").update(soul25).digest("hex");
if (soul36.parent_hash !== parentHash) throw new Error(`SOUL36_PARENT_HASH_MISMATCH:expected=${parentHash}:actual=${soul36.parent_hash}`);
if (soul36.nodes?.length !== 36 || soul36.counts?.total_nodes !== 36) throw new Error("SOUL36_NODE_COUNT_INVALID");
if (soul36.edges?.length !== 178 || soul36.counts?.edges !== 178) throw new Error("SOUL36_EDGE_COUNT_INVALID");

const oldNodeIds = new Set(JSON.parse(Buffer.from(soul25).toString("utf8")).nodes.map(x => x.id));
const missing = [...oldNodeIds].filter(id => !soul36.nodes.some(x => x.id === id));
if (missing.length) throw new Error(`SOUL25_NODES_NOT_PRESERVED:${missing.join(",")}`);

for (const id of ["bijux-dag-runtime","ouro-loop","recurs","fedml","hivemind","temporal","hora-graph-core","cognitive-workspace","ravana","ray","nervo-vago"]) {
  const node = soul36.nodes.find(x => x.id === id);
  if (!node) throw new Error(`SOUL36_NODE_MISSING:${id}`);
  if (node.state !== "PROJECTED") throw new Error(`SOUL36_EPISTEMIC_STATE_INVALID:${id}`);
}

const forbiddenSecondMesh = soul36.nodes.filter(x => /mesh/i.test(x.id) && x.id !== "nervo-vago");
if (forbiddenSecondMesh.length) throw new Error("SOUL36_SECOND_MESH_NODE_DETECTED");

console.log(JSON.stringify({
  ok:true,
  soul25_preserved:true,
  parent_hash:parentHash,
  soul36_nodes:36,
  soul36_edges:178,
  added_nodes:11,
  second_mesh:false,
  external_runtime_state:"PROJECTED"
},null,2));
