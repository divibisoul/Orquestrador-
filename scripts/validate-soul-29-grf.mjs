import { readFile } from 'node:fs/promises';
import { createHash } from 'node:crypto';

const raw25=await readFile('integrations/soul-25-capability-fabric.json');
const preserved25=await readFile('_preserved/soul-25-capability-fabric.json');
const soul25=JSON.parse(raw25);
if(!raw25.equals?.(preserved25)) {
  if(Buffer.from(raw25).compare(Buffer.from(preserved25))!==0) throw new Error('SOUL25_PRESERVED_COPY_NOT_BYTE_IDENTICAL');
}
const soul29=JSON.parse(await readFile('integrations/soul-29-capability-fabric.json','utf8'));
const contracts=JSON.parse(await readFile('integrations/grf/soul-29-participant-contracts.json','utf8'));
const binding=JSON.parse(await readFile('integrations/grf/system-binding.json','utf8'));
const invariantCatalog=JSON.parse(await readFile('integrations/grf/invariants.json','utf8'));
const nervoVago=JSON.parse(await readFile('integrations/grf/nervo-vago-unified-contract.json','utf8'));

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
const nervoNode=soul29.nodes.find(x=>x.id==='nervo-vago');
if(nervoNode.repository!=='https://github.com/OpenSIN-AI/OpenSIN-Neural-Bus') throw new Error('NERVOVAGO_SOURCE_ALIGNMENT_INVALID');
if(nervoNode.state!=='BLOCKED') throw new Error('NERVOVAGO_BLOCKED_STATE_INVALID');
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
if(!binding.unified_nervo_vago || binding.unified_nervo_vago.id!=='nervo-vago') throw new Error('NERVOVAGO_UNIFIED_BINDING_MISSING');
if(binding.unified_nervo_vago.no_second_vagus!==true || binding.unified_nervo_vago.no_second_mesh!==true) throw new Error('NERVOVAGO_UNIFICATION_POLICY_INVALID');
if(!nervoVago.component || nervoVago.component.id!=='nervo-vago') throw new Error('NERVOVAGO_CONTRACT_ID_INVALID');
if(nervoVago.authority?.no_second_vagus!==true || nervoVago.authority?.no_second_mesh!==true) throw new Error('NERVOVAGO_CONTRACT_POLICY_INVALID');
if(nervoVago.canonical_contract?.event_protocol!=='soul.vagus.event.v1') throw new Error('NERVOVAGO_EVENT_PROTOCOL_INVALID');
if(nervoVago.etr_gate?.required_before_delivery!==true) throw new Error('NERVOVAGO_ETR_GATE_INVALID');
const participants=contracts.contracts.map(x=>x.id).sort();
const expected=['autogenesis','belel-protocol','cognifold','cuda-oxide','functional-graph-agi','hora-graph-core','mycelium','octos','prime-agent','opensinn-bus'].sort();
if(JSON.stringify(participants)!==JSON.stringify(expected)) throw new Error('GRF_PARTICIPANT_SET_INVALID');
if(!binding.operations?.includes('nervo-vago.event@1.0.0') || !binding.operations?.includes('nervo-vago.describe@1.0.0')) {
  throw new Error('NERVOVAGO_OPERATIONS_MISSING');
}
if(nervoVago.execution_boundary?.implementation!=='backend.NervoVagoGateway') throw new Error('NERVOVAGO_EXECUTION_BOUNDARY_INVALID');
if(nervoVago.execution_boundary?.policy_gate!=='SARA/ETR via sara.audit@1.0.0') throw new Error('NERVOVAGO_POLICY_GATE_INVALID');
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
