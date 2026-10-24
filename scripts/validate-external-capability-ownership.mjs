import fs from 'node:fs/promises';

const registry = JSON.parse(await fs.readFile('integrations/external-capabilities.json', 'utf8'));
const ownership = JSON.parse(await fs.readFile('integrations/external-capability-ownership.json', 'utf8'));

const primary = registry.repositories ?? [];
const complementary = registry.complementary_repositories ?? [];
const expansion = registry.expansion_repositories ?? [];
const grce = registry.grce_complements ?? [];
const coreSources = [...primary, ...complementary];
const projectedSources = [...expansion, ...grce];
const owners = new Set(['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07']);
const sorted = values => [...values].sort();
const sourceIdSet = sources => sources.map(source => source.id).sort();
const coreRows = ownership.capabilities ?? [];
const projectedRows = ownership.expansion_capabilities ?? [];

if (coreSources.length !== 22 || projectedSources.length !== 10) {
  throw new Error(`SOURCE_CLASSIFICATION_COUNTS:core=${coreSources.length}:projected=${projectedSources.length}`);
}
if (ownership.counts?.primary_and_complementary !== 22 ||
    ownership.counts?.projected_expansions !== 10 ||
    ownership.counts?.total !== 32) {
  throw new Error('OWNERSHIP_DECLARED_COUNTS_MISMATCH');
}
if (JSON.stringify(sorted(sourceIdSet(coreSources))) !== JSON.stringify(sorted(sourceIdSet(coreRows)))) {
  throw new Error('CORE_OWNERSHIP_IDS_DO_NOT_MATCH_CORE_SOURCE_REGISTRY');
}
if (JSON.stringify(sorted(sourceIdSet(projectedSources))) !== JSON.stringify(sorted(sourceIdSet(projectedRows)))) {
  throw new Error('EXPANSION_OWNERSHIP_IDS_DO_NOT_MATCH_EXPANSION_REGISTRY');
}
const allRowIds = [...sourceIdSet(coreRows), ...sourceIdSet(projectedRows)];
if (new Set(allRowIds).size !== 32) throw new Error('OWNERSHIP_ID_COLLISION_OR_COUNT');
const sourceById = new Map([...coreSources, ...projectedSources].map(source => [source.id, source]));

function validateRow(row, projected) {
  if (!owners.has(row.primaryOwner)) throw new Error(`INVALID_PRIMARY_OWNER:${row.id}`);
  if (!Array.isArray(row.consumers) || row.consumers.includes(row.primaryOwner)) {
    throw new Error(`OWNER_MUST_BE_DISTINCT_FROM_CONSUMER:${row.id}`);
  }
  if (!['structural-only', 'adapter-planned', 'adapter-live'].includes(row.status)) {
    throw new Error(`INVALID_STATUS:${row.id}`);
  }
  const source = sourceById.get(row.id);
  if (!source) throw new Error(`SOURCE_NOT_FOUND:${row.id}`);
  const normalizedSource = String(source.source || '').replace(/\.git$/, '');
  const expectedSource = `https://github.com/${row.source}`;
  if (normalizedSource !== expectedSource) throw new Error(`SOURCE_MISMATCH:${row.id}`);
  const targets = source.targets ?? (source.target ? [source.target] : []);
  if (!targets.includes(row.primaryOwner)) throw new Error(`PRIMARY_OWNER_NOT_A_DECLARED_TARGET:${row.id}`);
  if (projected) {
    if (row.mode !== 'expansion') throw new Error(`PROJECTED_OWNERSHIP_MODE_INVALID:${row.id}`);
    if (row.status === 'adapter-live') throw new Error(`PROJECTED_OWNERSHIP_PREMATURELY_LIVE:${row.id}`);
  } else if (row.mode === 'expansion') {
    throw new Error(`CORE_MARKED_AS_EXPANSION:${row.id}`);
  }
}
for (const row of coreRows) validateRow(row, false);
for (const row of projectedRows) validateRow(row, true);

console.log(JSON.stringify({
  ok: true,
  primary_sources: primary.length,
  complementary_adapter_sources: complementary.length,
  core_capabilities: coreRows.length,
  projected_expansion_sources: projectedSources.length,
  projected_expansion_ownership: projectedRows.length,
  total_sources: 32,
  owners: [...new Set([...coreRows, ...projectedRows].map(row => row.primaryOwner))].sort()
}, null, 2));
