import fs from 'node:fs/promises';

const registry = JSON.parse(await fs.readFile('integrations/external-capabilities.json','utf8'));
const ownership = JSON.parse(await fs.readFile('integrations/external-capability-ownership.json','utf8'));

const registryIds = registry.repositories.map(x=>x.id).sort();
const rows = ownership.capabilities;
if (rows.length !== 16) throw new Error(`OWNERSHIP_COUNT:${rows.length}`);
const ids = rows.map(x=>x.id).sort();
if (JSON.stringify(ids) !== JSON.stringify(registryIds)) throw new Error('OWNERSHIP_IDS_DO_NOT_MATCH_REGISTRY');
const owners = new Set(['N01','N02','N03','N04','N05','N06','N07']);
for (const row of rows) {
  if (!owners.has(row.primaryOwner)) throw new Error(`INVALID_PRIMARY_OWNER:${row.id}`);
  if (!Array.isArray(row.consumers) || row.consumers.includes(row.primaryOwner)) throw new Error(`OWNER_MUST_BE_DISTINCT_FROM_CONSUMER:${row.id}`);
  if (!['structural-only','adapter-planned','adapter-live'].includes(row.status)) throw new Error(`INVALID_STATUS:${row.id}`);
  const source = registry.repositories.find(x=>x.id===row.id);
  if (!source || source.source !== `https://github.com/${row.source}.git` && source.source !== `https://github.com/${row.source}`) {
    throw new Error(`SOURCE_MISMATCH:${row.id}`);
  }
}

const targetMap = new Map(registry.repositories.map(x => [x.id, new Set(x.targets)]));
for (const row of rows) {
  const targets = targetMap.get(row.id);
  if (!targets?.has(row.primaryOwner)) throw new Error(`PRIMARY_OWNER_NOT_A_DIRECT_TARGET:${row.id}`);
  if (row.id === 'superpowers' && row.mode !== 'dev-skill-pack') throw new Error('SUPERPOWERS_MODE_INVALID');
}

console.log(JSON.stringify({ok:true,capabilities:rows.length,owners:[...new Set(rows.map(x=>x.primaryOwner))].sort()},null,2));
