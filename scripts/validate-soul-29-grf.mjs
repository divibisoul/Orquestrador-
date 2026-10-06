import { readFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';

const raw25=await readFile('integrations/soul-25-capability-fabric.json');
const soul25=JSON.parse(raw25);
const soul29=JSON.parse(await readFile('integrations/soul-29-capability-fabric.json','utf8'));
const contracts=JSON.parse(await readFile('integrations/grf/soul-29-participant-contracts.json','utf8'));
const binding=JSON.parse(await readFile('integrations/grf/system-binding.json','utf8'));
const invariantCatalog=JSON.parse(await readFile('integrations/grf/invariants.json','utf8'));

const parentHash=createHash('sha256').update(raw25).digest('hex');
if(soul29.parent_hash!==`sha256:${parentHash}`) throw new Error(`SOUL29_PARENT_HASH_MISMATCH:expected=sha256:${parentHash}:actual=${soul29.parent_hash}`);
if(soul29.version!==29) throw new Error(`SOUL29_VERSION:${soul29.version}`);
if(soul29.nodes.length!==29 || soul29.counts?.total_nodes!==29) throw new Error('SOUL29_NODE_COUNT_INVALID');
if(soul29.edges.length!==160 || soul29.counts?.edges!==160) throw new Error('SOUL29_EDGE_COUNT_INVALID');

const missingOld=soul25.nodes.filter(x=>!soul29.nodes.some(y=>y.id===x.id));
if(missingOld.length) throw new Error(`SOUL25_NODES_NOT_PRESERVED:${missingOld.map(x=>x.id).join(',')}`);
if(soul29.nodes.length-soul25.nodes.length!==4) throw new Error('SOUL29_NEW_NODE_DELTA_INVALID');
if(soul29.edges.length-soul25.edges.length!==16) throw new Error('SOUL29_NEW_EDGE_DELTA_INVALID');
for(const edge of soul25.edges){
  if(!soul29.edges.some(candidate=>JSON.stringify(candidate)===JSON.stringify(edge))) throw new Error(`SOUL25_EDGE_NOT_PRESERVED:${JSON.stringify(edge)}`);
}
for(const id of ['autogenesis','supergpu-agi','clareira-agi','nervo-vago']){
  if(!soul29.nodes.some(x=>x.id===id)) throw new Error(`SOUL29_NODE_MISSING:${id}`);
}
if(contracts.contracts?.length!==10 || contracts.contracts_count!==10) throw new Error('GRF_NEW_CONTRACT_COUNT_INVALID');
const required=['ingest','epistemicState','invariants','capabilities','failuresAbsorbed','provenance'];
for(const c of contracts.contracts){
  for(const method of required) if(!c.interface?.includes(method)) throw new Error(`GRF_PARTICIPANT_INTERFACE_MISSING:${c.id}:${method}`);
  if(!['PROJECTED','BLOCKED','REAL'].includes(c.state)) throw new Error(`GRF_PARTICIPANT_STATE_INVALID:${c.id}`);
}
const invariantIds=['I1','I2','I3','I4','I5','I6','I7','I8','I9','I10','I11','I12','I13','I14','I15','I16','I17','I18'];
if(invariantCatalog.count!==18 || JSON.stringify(invariantCatalog.invariants.map(x=>x.id))!==JSON.stringify(invariantIds)) throw new Error('GRF_INVARIANT_CATALOG_INVALID');
const bindingLayers=new Set((binding.layer_contracts||[]).map(x=>x.id));
for(const id of ['L0','L1','L2','L3','L4','L5','L6','L7']) if(!bindingLayers.has(id)) throw new Error(`GRF_LAYER_MISSING:${id}`);
const participants=contracts.contracts.map(x=>x.id).sort();
const expected=['autogenesis','belel-protocol','cognifold','cuda-oxide','functional-graph-agi','hora-graph-core','mycelium','octos','prime-agent','opensinn-bus'].sort();
if(JSON.stringify(participants)!==JSON.stringify(expected)) throw new Error('GRF_PARTICIPANT_SET_INVALID');
console.log(JSON.stringify({
  ok:true,
  soul25_nodes:soul25.nodes.length,
  soul29_nodes:soul29.nodes.length,
  soul25_preserved:true,
  soul29_edges:soul29.edges.length,
  grf_invariants:invariantCatalog.count,
  grf_layers:8,
  new_participant_contracts:10
},null,2));
