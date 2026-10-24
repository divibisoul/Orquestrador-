import fs from 'node:fs/promises';

const registry=JSON.parse(await fs.readFile('integrations/external-capabilities.json','utf8'));
const adapters=JSON.parse(await fs.readFile('integrations/external-adapters.json','utf8'));
const ownership=JSON.parse(await fs.readFile('integrations/external-capability-ownership.json','utf8'));
const canonical=new Set(['N01','N02','N03','N04','N05','N06','N07']);

const primarySources=registry.repositories ?? [];
const complementarySources=registry.complementary_repositories ?? [];
const allSources=[...primarySources,...complementarySources];
const sourceIds=allSources.map(x=>x.id).sort();
const adapterIds=(adapters.providers ?? []).map(x=>x.id).sort();
const ownerIds=(ownership.capabilities ?? []).map(x=>x.id).sort();

if(primarySources.length!==16) throw new Error(`PRIMARY_REGISTRY_COUNT:${primarySources.length}`);
if(complementarySources.length!==6) throw new Error(`COMPLEMENTARY_REGISTRY_COUNT:${complementarySources.length}`);
if(sourceIds.length!==22) throw new Error(`TOTAL_ACTIVE_REGISTRY_COUNT:${sourceIds.length}`);
if(JSON.stringify(sourceIds)!==JSON.stringify(adapterIds)) throw new Error('ADAPTER_IDS_DO_NOT_MATCH_22_ACTIVE_SOURCE_REGISTRY');
if(JSON.stringify(sourceIds)!==JSON.stringify(ownerIds)) throw new Error('OWNERSHIP_IDS_DO_NOT_MATCH_22_ACTIVE_SOURCE_REGISTRY');
if(adapters.canonical_control_plane!=='N07') throw new Error('ADAPTER_CONTROL_PLANE_INVALID');

for(const p of adapters.providers ?? []){
  if(!canonical.has(p.owner)) throw new Error(`INVALID_OWNER:${p.id}`);
  if(p.root!==`integrations/external/${p.id}`) throw new Error(`NON_CANONICAL_ROOT:${p.id}`);
  if(!Array.isArray(p.operations)||p.operations.length===0) throw new Error(`NO_OPERATIONS:${p.id}`);
  const src=allSources.find(x=>x.id===p.id);
  if(!src || !src.targets.includes(p.owner)) throw new Error(`OWNER_NOT_DIRECT_TARGET:${p.id}`);
}

// Expansion and GRCE complements are visible to federation discovery, but are
// not runtime-routeable until a native adapter is built and independently proven.
const activeIds=new Set(sourceIds);
const expansionRegistry=[...(registry.expansion_repositories ?? []),...(registry.grce_complements ?? [])];
const projectedSourceByID=new Map();
for(const src of expansionRegistry){
  if(activeIds.has(src.id)) continue; // hora-graph-core is already an active complementary source.
  const prior=projectedSourceByID.get(src.id);
  if(prior && (prior.revision!==src.revision || prior.source!==src.source)) {
    throw new Error(`CONFLICTING_PROJECTED_SOURCE:${src.id}`);
  }
  projectedSourceByID.set(src.id,src);
}
const projectedSources=[...projectedSourceByID.values()].sort((a,b)=>a.id.localeCompare(b.id));
const expansionProviders=adapters.expansion_providers ?? [];
const expansionOwnership=ownership.expansion_capabilities ?? [];
const projectedIDs=projectedSources.map(x=>x.id).sort();
const expansionIDs=expansionProviders.map(x=>x.id).sort();
const expansionOwnerIDs=expansionOwnership.map(x=>x.id).sort();
if(projectedSources.length!==9) throw new Error(`PROJECTED_SOURCE_COUNT:${projectedSources.length}`);
if(JSON.stringify(projectedIDs)!==JSON.stringify(expansionIDs)) throw new Error('EXPANSION_ADAPTER_IDS_DO_NOT_MATCH_PROJECTED_SOURCES');
if(JSON.stringify(projectedIDs)!==JSON.stringify(expansionOwnerIDs)) throw new Error('EXPANSION_OWNERSHIP_IDS_DO_NOT_MATCH_PROJECTED_SOURCES');

const normRepo=(value)=>String(value ?? '').replace(/^https:\/\/github\.com\//,'').replace(/\.git$/,'').replace(/\/$/,'');
for(const p of expansionProviders){
  if(activeIds.has(p.id)) throw new Error(`PROJECTED_PROVIDER_DUPLICATES_ACTIVE_PROVIDER:${p.id}`);
  if(!canonical.has(p.owner)) throw new Error(`INVALID_EXPANSION_OWNER:${p.id}`);
  if(p.state!=='PROJECTED') throw new Error(`EXPANSION_MUST_REMAIN_PROJECTED:${p.id}:${p.state}`);
  if(p.root!==`integrations/external/${p.id}`) throw new Error(`NON_CANONICAL_EXPANSION_ROOT:${p.id}`);
  if(!Array.isArray(p.operations)||p.operations.length===0) throw new Error(`NO_EXPANSION_OPERATIONS:${p.id}`);
  if(!p.integration_state || !p.runtime_requirement) throw new Error(`EXPANSION_EVIDENCE_GATES_MISSING:${p.id}`);
  const src=projectedSourceByID.get(p.id);
  const targetList=[...(src?.targets ?? []),...(src?.target ? [src.target] : [])];
  if(!src || !targetList.includes(p.owner)) throw new Error(`EXPANSION_OWNER_NOT_DECLARED_AS_TARGET:${p.id}`);
  const own=expansionOwnership.find(x=>x.id===p.id);
  if(!own || own.primaryOwner!==p.owner) throw new Error(`EXPANSION_OWNER_CONTRACT_MISMATCH:${p.id}`);
  if(own.pinnedRevision!==src.revision) throw new Error(`EXPANSION_PIN_MISMATCH:${p.id}`);
  if(normRepo(own.source)!==normRepo(src.source)) throw new Error(`EXPANSION_SOURCE_URL_MISMATCH:${p.id}`);
  if(own.status!=='projected-no-runtime-route') throw new Error(`EXPANSION_OWNERSHIP_MUST_REMAIN_PROJECTED:${p.id}`);
  if(!Array.isArray(own.consumers) || nucCount(own.consumers)!==7) throw new Error(`EXPANSION_MUST_BE_DISCOVERABLE_FROM_ALL_SEVEN_NUCLEI:${p.id}`);
  for(const consumer of own.consumers){
    if(!canonical.has(consumer)) throw new Error(`INVALID_EXPANSION_CONSUMER:${p.id}:${consumer}`);
  }
}
function nucCount(values){return new Set(values.filter(v=>canonical.has(v))).size;}

console.log(JSON.stringify({
  ok:true,
  primary_upstreams:primarySources.length,
  complementary_sources:complementarySources.length,
  active_adapters:sourceIds.length,
  projected_expansions:projectedSources.length,
  total_pinned_external_repositories:sourceIds.length+projectedSources.length,
  control_plane:'N07',
  execution_policy:'projected expansions are discoverable but not routeable'
},null,2));
