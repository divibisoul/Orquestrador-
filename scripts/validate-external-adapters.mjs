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
if(complementarySources.length!==6) throw new Error(`COMPLEMENTARY_REGISTRY_COUNT:${complementarySources.length}`);
if(sourceIds.length!==22) throw new Error(`TOTAL_REGISTRY_COUNT:${sourceIds.length}`);
if(JSON.stringify(sourceIds)!==JSON.stringify(adapterIds)) throw new Error('ADAPTER_IDS_DO_NOT_MATCH_22_SOURCE_REGISTRY');
if(JSON.stringify(sourceIds)!==JSON.stringify(ownerIds)) throw new Error('OWNERSHIP_IDS_DO_NOT_MATCH_22_SOURCE_REGISTRY');
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
  complementary_sources:6,
  total_sources:22,
  adapter_roots:'integrations/external/*',
  control_plane:'N07'
},null,2));