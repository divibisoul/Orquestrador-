import fs from 'node:fs/promises';

const registry=JSON.parse(await fs.readFile('integrations/external-capabilities.json','utf8'));
const adapters=JSON.parse(await fs.readFile('integrations/external-adapters.json','utf8'));
const ownership=JSON.parse(await fs.readFile('integrations/external-capability-ownership.json','utf8'));
const canonical=new Set(['N01','N02','N03','N04','N05','N06','N07']);

const primarySources=registry.repositories ?? [];
const complementarySources=registry.complementary_repositories ?? [];
const allSources=[...primarySources,...complementarySources];
const sourceIds=allSources.map(x=>x.id).sort();
const adapterIds=adapters.providers.map(x=>x.id).sort();
const ownerIds=ownership.capabilities.map(x=>x.id).sort();

if(primarySources.length!==16) throw new Error(`PRIMARY_REGISTRY_COUNT:${primarySources.length}`);
// Preserve the historical six-source complementary set while accepting the expanded, pinned
// 16-source set added by the active public-capability recovery front.
const historicalComplementaryCount=6;
const expandedComplementaryIds=[
  'autogenesis','octos','hora-graph-core','mycelium','prime-agent','cuda-oxide',
  'bijux-dag-runtime','ouro-loop','recuris','fedml','hivemind','temporal',
  'cognitive-workspace','ravana','ray','nats-go'
].sort();
if(complementarySources.length!==historicalComplementaryCount && complementarySources.length!==expandedComplementaryIds.length)
  throw new Error(`COMPLEMENTARY_REGISTRY_COUNT:${complementarySources.length}`);
if(sourceIds.length!==primarySources.length+complementarySources.length)
  throw new Error(`TOTAL_REGISTRY_COUNT_MISMATCH:${sourceIds.length}`);
if(new Set(sourceIds).size!==sourceIds.length) throw new Error('DUPLICATE_SOURCE_IDS_IN_REGISTRY');
if(complementarySources.length===expandedComplementaryIds.length) {
  const actualComplementaryIds=complementarySources.map(x=>x.id).sort();
  if(JSON.stringify(actualComplementaryIds)!==JSON.stringify(expandedComplementaryIds))
    throw new Error('EXPANDED_COMPLEMENTARY_SOURCE_SET_MISMATCH');
}
for(const source of allSources) {
  if(!/^[a-f0-9]{40}$/i.test(String(source.revision||'')))
    throw new Error(`SOURCE_REVISION_NOT_PINNED_TO_COMMIT:${source.id}`);
}
if(JSON.stringify(sourceIds)!==JSON.stringify(adapterIds)) throw new Error('ADAPTER_IDS_DO_NOT_MATCH_SOURCE_REGISTRY');
if(JSON.stringify(sourceIds)!==JSON.stringify(ownerIds)) throw new Error('OWNERSHIP_IDS_DO_NOT_MATCH_SOURCE_REGISTRY');
if(adapters.canonical_control_plane!=='N07') throw new Error('ADAPTER_CONTROL_PLANE_INVALID');

for(const p of adapters.providers){
  if(!canonical.has(p.owner)) throw new Error(`INVALID_OWNER:${p.id}`);
  if(!p.root.startsWith('integrations/external/')) throw new Error(`NON_CANONICAL_ROOT:${p.id}`);
  if(!Array.isArray(p.operations)||p.operations.length===0) throw new Error(`NO_OPERATIONS:${p.id}`);
  const src=allSources.find(x=>x.id===p.id);
  if(!src || !src.targets.includes(p.owner)) throw new Error(`OWNER_NOT_DIRECT_TARGET:${p.id}`);
}

console.log(JSON.stringify({
  ok:true,
  primary_upstreams:16,
  complementary_sources:complementarySources.length,
  total_sources:allSources.length,
  adapter_roots:'integrations/external/*',
  control_plane:'N07'
},null,2));