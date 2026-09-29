import assert from 'node:assert/strict';
import fs from 'node:fs';

const manifest = JSON.parse(fs.readFileSync(new URL('../soul-nuclei.json', import.meta.url), 'utf8'));
assert.equal(manifest.system, 'SOUL');
assert.ok(manifest.primordialEssence);
assert.equal(manifest.primordialEssence.definition.includes('not a claim about exact historical creation date'), true);

const expected = ['N01','N02','N03','N04','N05','N06','N07','SARA'];
assert.deepEqual(Object.keys(manifest.primordialEssence.essences).sort(), expected.sort());

for (const id of expected) {
  const item = manifest.primordialEssence.essences[id];
  assert.ok(item.nativeRole);
  assert.ok(item.essence);
  assert.ok(Array.isArray(item.evidence) && item.evidence.length > 0);
}

assert.ok(Array.isArray(manifest.primordialEssence.compositionSeeds));
for (const seed of manifest.primordialEssence.compositionSeeds) {
  assert.ok(seed.participants.length >= 2);
  assert.ok(seed.derivedFunction);
  assert.ok(seed.mode);
  assert.ok(seed.existingEvidence.length > 0);
  assert.equal(['STRUCTURAL','EXECUTABLE','VERIFIED','FUSED','BLOCKED'].includes(seed.status), true);
}

const directAdjacent = manifest.primordialEssence.compositionSeeds.filter(s => s.mode === 'adjacent');
assert.equal(directAdjacent.length >= 6, true);

console.log(JSON.stringify({
  status:'PASS',
  continuityLedgerVersion:manifest.continuityLedgerVersion,
  essences:expected.length,
  compositionSeeds:manifest.primordialEssence.compositionSeeds.length,
  adjacentSeeds:directAdjacent.length
},null,2));
