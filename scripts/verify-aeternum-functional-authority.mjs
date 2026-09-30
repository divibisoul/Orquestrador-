import assert from 'node:assert/strict';

const SOURCE = 'https://raw.githubusercontent.com/divibisoul/aeternum-core-29/eb9e39134896903b86f3fc36a2ee679971c8c8c3/docs/aeternum-functional-authority.snapshot.json';

const response = await fetch(SOURCE);
if (!response.ok) throw new Error('AETERNUM_AUTHORITY_SOURCE_HTTP_' + response.status);
const snapshot = await response.json();

assert.equal(snapshot.schemaVersion, '1.0.0');
assert.equal(snapshot.system, 'SOUL');
assert.equal(snapshot.layer, 'AETERNUM');
assert.equal(snapshot.authorityModel, 'connection-derived');
assert.equal(snapshot.fusion?.protocol, 'nvod-fusion/1');
assert.equal(snapshot.fusion?.contractVersion, '1.0.0');
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
  source: SOURCE,
  moduleCount: snapshot.modules.length,
  anchor: 'M1_CORE',
  anchorConnectionScore: m1.connectionScore,
}));
