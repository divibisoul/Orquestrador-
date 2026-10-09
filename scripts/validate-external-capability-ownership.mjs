import fs from 'node:fs/promises';

const registry = JSON.parse(await fs.readFile('integrations/external-capabilities.json','utf8'));
const ownership = JSON.parse(await fs.readFile('integrations/external-capability-ownership.json','utf8'));

const primary = registry.repositories ?? [];
const complementary = registry.complementary_repositories ?? [];
const allSources = [...primary, ...complementary];
const registryIds = allSources.map(x=>x.id).sort();
const rows = ownership.capabilities;
if (rows.length !== allSources.length) throw new Error(`OWNERSHIP_COUNT:${rows.length}:EXPECTED:${allSources.length}`);
const ids = rows.map(x=>x.id).sort();
if (JSON.stringify(ids) !== JSON.stringify(registryIds)) throw new Error('OWNERSHIP_IDS_DO_NOT_MATCH_SOURCE_REGISTRY');
const owners = new Set(['N01','N02','N03','N04','N05','N06','N07']);
for (const row of rows) {
  if (!owners.has(row.primaryOwner)) throw new Error(`INVALID_PRIMARY_OWNER:${row.id}`);
  if (!Array.isArray(row.consumers) || row.consumers.includes(row.primaryOwner)) throw new Error(`OWNER_MUST_BE_DISTINCT_FROM_CONSUMER:${row.id}`);
  if (!['structural-only','adapter-planned','adapter-live'].includes(row.status)) throw new Error(`INVALID_STATUS:${row.id}`);
  const source = allSources.find(x=>x.id===row.id);
  if (!source) throw new Error(`SOURCE_NOT_FOUND:${row.id}`);
  const normalizedSource = String(source.source || '').replace(/\.git$/,'');
  const expectedSource = `https://github.com/${row.source}`;
  if (normalizedSource !== expectedSource) throw new Error(`SOURCE_MISMATCH:${row.id}`);
  if (!Array.isArray(source.targets) || !source.targets.includes(row.primaryOwner)) throw new Error(`PRIMARY_OWNER_NOT_A_DIRECT_TARGET:${row.id}`);
}
const primaryIds = new Set(primary.map(x=>x.id));
const complementaryIds = new Set(complementary.map(x=>x.id));
for (const row of rows) {
  if (complementaryIds.has(row.id) && row.mode !== 'complementary') throw new Error(`COMPLEMENTARY_MODE_INVALID:${row.id}`);
  if (primaryIds.has(row.id) && row.mode === 'complementary') throw new Error(`PRIMARY_MARKED_COMPLEMENTARY:${row.id}`);
}
console.log(JSON.stringify({
  ok:true,
  primary_sources:primary.length,
  complementary_sources:complementary.length,
  capabilities:rows.length,
  owners:[...new Set(rows.map(x=>x.primaryOwner))].sort()
},null,2));