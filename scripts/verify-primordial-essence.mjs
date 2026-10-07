import assert from 'node:assert/strict';
import fs from 'node:fs';

const manifest = JSON.parse(fs.readFileSync(new URL('../soul-nuclei.json', import.meta.url), 'utf8'));
assert.equal(manifest.system, 'SOUL');
assert.ok(manifest.primordialEssence);
assert.equal(
  manifest.primordialEssence.definition.includes('not a claim about exact historical creation date'),
  true
);

const expected = ['N01','N02','N03','N04','N05','N06','N07','SARA'];
assert.deepEqual(Object.keys(manifest.primordialEssence.essences).sort(), expected.slice().sort());

for (const id of expected) {
  const item = manifest.primordialEssence.essences[id];
  assert.ok(item.nativeRole);
  assert.ok(item.essence);
  assert.ok(Array.isArray(item.evidence) && item.evidence.length > 0);
}

assert.ok(Array.isArray(manifest.primordialEssence.compositionSeeds));
for (const seed of manifest.primordialEssence.compositionSeeds) {
  assert.ok(Array.isArray(seed.participants) && seed.participants.length >= 2);
  assert.ok(seed.derivedFunction);
  assert.ok(seed.mode);
  assert.ok(Array.isArray(seed.existingEvidence) && seed.existingEvidence.length > 0);
  assert.ok(['STRUCTURAL','EXECUTABLE','VERIFIED','FUSED','BLOCKED'].includes(seed.status));
}

const adjacent = manifest.primordialEssence.compositionSeeds.filter(seed => seed.mode === 'adjacent');
assert.ok(adjacent.length >= 6);

console.log(JSON.stringify({
  status: 'PASS',
  continuityLedgerVersion: manifest.continuityLedgerVersion,
  essences: expected.length,
  compositionSeeds: manifest.primordialEssence.compositionSeeds.length,
  adjacentSeeds: adjacent.length
}, null, 2));
