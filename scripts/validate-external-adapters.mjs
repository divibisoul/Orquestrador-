import fs from 'node:fs/promises';

const registry = JSON.parse(await fs.readFile('integrations/external-capabilities.json', 'utf8'));
const adapters = JSON.parse(await fs.readFile('integrations/external-adapters.json', 'utf8'));
const ownership = JSON.parse(await fs.readFile('integrations/external-capability-ownership.json', 'utf8'));

const primary = registry.repositories ?? [];
const complementary = registry.complementary_repositories ?? [];
const expansion = registry.expansion_repositories ?? [];
const grce = registry.grce_complements ?? [];
const coreSources = [...primary, ...complementary];
const projectedSources = [...expansion, ...grce];
const allSources = [...coreSources, ...projectedSources];
const canonicalOwners = new Set(['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07']);
const sorted = values => [...values].sort();
const idsOf = values => values.map(value => value.id).sort();

if (primary.length !== 16) throw new Error(`PRIMARY_REGISTRY_COUNT:${primary.length}`);
if (complementary.length !== 6) throw new Error(`COMPLEMENTARY_REGISTRY_COUNT:${complementary.length}`);
if (coreSources.length !== 22) throw new Error(`CORE_SOURCE_COUNT:${coreSources.length}`);
if (projectedSources.length !== 10) throw new Error(`PROJECTED_EXPANSION_COUNT:${projectedSources.length}`);
if (allSources.length !== 32) throw new Error(`TOTAL_PINNED_SOURCE_COUNT:${allSources.length}`);
const allIds = idsOf(allSources);
if (new Set(allIds).size !== allIds.length) throw new Error('DUPLICATE_SOURCE_IDS_IN_REGISTRY');

const coreIds = idsOf(coreSources);
const projectedIds = idsOf(projectedSources);
const adapterIds = idsOf(adapters.providers ?? []);
const projectedAdapterIds = idsOf(adapters.projected_expansions ?? []);
if (adapters.canonical_control_plane !== 'N07') throw new Error('ADAPTER_CONTROL_PLANE_INVALID');
if (adapters.provider_count !== 22 || adapters.providers?.length !== 22) {
  throw new Error(`ADAPTER_BACKED_PROVIDER_COUNT:${adapters.providers?.length}:declared:${adapters.provider_count}`);
}
if (JSON.stringify(sorted(coreIds)) !== JSON.stringify(sorted(adapterIds))) {
  throw new Error('EXECUTABLE_ADAPTER_IDS_DO_NOT_MATCH_CORE_SOURCE_REGISTRY');
}
if (adapters.projected_expansion_count !== 10 || adapters.projected_expansions?.length !== 10) {
  throw new Error(`PROJECTED_ADAPTER_COUNT:${adapters.projected_expansions?.length}:declared:${adapters.projected_expansion_count}`);
}
if (JSON.stringify(sorted(projectedIds)) !== JSON.stringify(sorted(projectedAdapterIds))) {
  throw new Error('PROJECTED_ADAPTER_IDS_DO_NOT_MATCH_EXPANSION_REGISTRY');
}
const ownershipCoreIds = idsOf(ownership.capabilities ?? []);
const ownershipExpansionIds = idsOf(ownership.expansion_capabilities ?? []);
if (JSON.stringify(sorted(coreIds)) !== JSON.stringify(sorted(ownershipCoreIds))) throw new Error('CORE_OWNERSHIP_IDS_MISMATCH');
if (JSON.stringify(sorted(projectedIds)) !== JSON.stringify(sorted(ownershipExpansionIds))) throw new Error('EXPANSION_OWNERSHIP_IDS_MISMATCH');
if (new Set([...ownershipCoreIds, ...ownershipExpansionIds]).size !== 32) throw new Error('OWNERSHIP_SOURCE_COUNT_NOT_32');

const revisions = new Map();
for (const source of allSources) {
  if (!/^[a-f0-9]{40}$/i.test(String(source.revision ?? ''))) {
    throw new Error(`SOURCE_REVISION_NOT_PINNED_TO_COMMIT:${source.id}`);
  }
  revisions.set(source.id, source);
}

for (const p of adapters.providers) {
  if (!canonicalOwners.has(p.owner)) throw new Error(`INVALID_CORE_OWNER:${p.id}`);
  if (p.root !== `integrations/external/${p.id}`) throw new Error(`NON_CANONICAL_ROOT:${p.id}`);
  if (!Array.isArray(p.operations) || p.operations.length === 0) throw new Error(`NO_CORE_OPERATIONS:${p.id}`);
  const source = revisions.get(p.id);
  if (!source || !coreIds.includes(p.id)) throw new Error(`CORE_SOURCE_NOT_FOUND:${p.id}`);
  if (!Array.isArray(source.targets) || !source.targets.includes(p.owner)) {
    throw new Error(`CORE_OWNER_NOT_A_DIRECT_TARGET:${p.id}`);
  }
}

for (const p of adapters.projected_expansions) {
  if (!canonicalOwners.has(p.owner)) throw new Error(`INVALID_PROJECTED_OWNER:${p.id}`);
  if (p.root !== `integrations/external/${p.id}`) throw new Error(`PROJECTED_ROOT_NOT_CANONICAL:${p.id}`);
  if (p.state !== 'PROJECTED') throw new Error(`PROJECTED_SOURCE_PROMOTED:${p.id}`);
  if (p.execution_enabled !== false || p.execution_proven !== false) throw new Error(`PROJECTED_SOURCE_EXECUTION_OVERCLAIM:${p.id}`);
  if (!Array.isArray(p.runtime_route_operations) || p.runtime_route_operations.length !== 0) {
    throw new Error(`PROJECTED_SOURCE_HAS_RUNTIME_ROUTE:${p.id}`);
  }
  const metadata = new Set(p.metadata_operations ?? []);
  if (!metadata.has(`external.${p.id}.describe@1.0.0`) || !metadata.has(`external.${p.id}.probe@1.0.0`)) {
    throw new Error(`PROJECTED_METADATA_OPERATIONS_MISSING:${p.id}`);
  }
  const source = revisions.get(p.id);
  if (!source) throw new Error(`PROJECTED_SOURCE_NOT_FOUND:${p.id}`);
  const targets = source.targets ?? (source.target ? [source.target] : []);
  if (!targets.includes(p.owner)) throw new Error(`PROJECTED_OWNER_NOT_A_DECLARED_TARGET:${p.id}`);
}

console.log(JSON.stringify({
  ok: true,
  primary_sources: primary.length,
  complementary_adapter_sources: complementary.length,
  adapter_backed_sources: coreSources.length,
  projected_expansions: projectedSources.length,
  total_unique_pinned_sources: allSources.length,
  executable_provider_count: adapters.providers.length,
  projected_metadata_only_count: adapters.projected_expansions.length,
  control_plane: 'N07',
  rule: 'only core adapters can execute; projected expansions remain discoverable, pinned, and fail-closed'
}, null, 2));
