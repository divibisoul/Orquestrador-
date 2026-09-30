import assert from 'node:assert/strict';

const fs = await import('node:fs/promises');
const localContract = JSON.parse(await fs.readFile(new URL('../contracts/aeternum-functional-authority.json', import.meta.url), 'utf8'));
const snapshot = localContract.snapshot;

assert.equal(localContract.fusion.nvodProtocol, 'nvod-fusion/1');
assert.equal(localContract.fusion.nvodContractVersion, '1.0.0');
assert.equal(snapshot.schemaVersion, '1.0.0');
assert.equal(snapshot.authorityModel, 'connection-derived');
assert.equal(snapshot.modules.length, 8);

const byId = new Map();
for (const module of snapshot.modules) {
  assert.equal(byId.has(module.moduleId), false, 'duplicate module: ' + module.moduleId);
  byId.set(module.moduleId, module);
}

for (const module of snapshot.modules) {
  for (const dependency of module.directDependencies) {
    assert.equal(byId.has(dependency), true, module.moduleId + ' missing dependency ' + dependency);
    assert.equal(
      byId.get(dependency).directDependents.includes(module.moduleId),
      true,
      'dependency inverse missing for ' + module.moduleId + ' -> ' + dependency,
    );
  }

  const expectedScore = (module.directDependents?.length ?? 0) * 2 + (module.transitiveDependents?.length ?? 0);
  assert.equal(module.connectionScore, expectedScore, 'connection score mismatch: ' + module.moduleId);

  const expectedPrefix = module.moduleId === 'M1_CORE'
    ? ['M1_CORE']
    : module.moduleId === 'M2_ORCHESTRATION'
      ? ['M1_CORE', 'M2_ORCHESTRATION']
      : module.authorityPath.slice(0, 1);
  assert.deepEqual(module.authorityPath.slice(0, expectedPrefix.length), expectedPrefix);

  assert.ok(module.executionOwner, 'execution owner missing: ' + module.moduleId);
}

const m1 = byId.get('M1_CORE');
assert.ok(m1.roles.includes('ANCHOR'));
assert.equal(m1.connectionScore, 15);

const m6 = byId.get('M6_IMMUNITY');
assert.equal(m6.governanceAuthority, 'SARA');
assert.equal(m6.executionOwner, 'N07');

const m8 = byId.get('M8_GOVERNANCE_MEMORY');
assert.equal(m8.executionOwner, 'N07');
assert.equal(m8.governanceAuthority, 'SARA');
assert.ok(m8.roles.includes('MEMORY'));

console.log(JSON.stringify({
  check: 'N07_AETERNUM_FUNCTIONAL_AUTHORITY_CROSSFRONT',
  status: 'PASS',
  source: localContract.source,
  moduleCount: snapshot.modules.length,
  anchor: 'M1_CORE',
  anchorConnectionScore: m1.connectionScore,
}));
