import fs from "node:fs";

const file="integrations/soul-28-capability-fabric.json";
const d=JSON.parse(fs.readFileSync(file,"utf8"));
if(d.counts.total_nodes!==28) throw new Error(`SOUL28_NODE_COUNT:${d.counts.total_nodes}`);
if(d.counts.edges!==156) throw new Error(`SOUL28_EDGE_COUNT:${d.counts.edges}`);
if(d.nodes.length!==28 || d.edges.length!==156) throw new Error("SOUL28_ARRAY_COUNTS_MISMATCH");
const ids=new Set(d.nodes.map(x=>x.id));
for(const e of d.edges){ if(!ids.has(e.from)||!ids.has(e.to)) throw new Error(`SOUL28_EDGE_ENDPOINT_MISSING:${e.from}->${e.to}`); }
const complements=["bijux-dag-runtime","ouro-loop","recurs"];
for(const id of complements){ if(!ids.has(id)) throw new Error(`SOUL28_COMPLEMENT_MISSING:${id}`); }
const counts={};
for(const e of d.edges){ if(complements.includes(e.from)) counts[e.from]=(counts[e.from]||0)+1; }
for(const id of complements){ if(counts[id]!==4) throw new Error(`SOUL28_EDGE_DEGREE:${id}:${counts[id]}`); }
if(!/^sha256:[0-9a-f]{64}$/.test(d.parent_hash)) throw new Error("SOUL28_PARENT_HASH_INVALID");
console.log(JSON.stringify({ok:true,nodes:d.nodes.length,edges:d.edges.length,parent_hash:d.parent_hash,complements,edge_degree:counts},null,2));