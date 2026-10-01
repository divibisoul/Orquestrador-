import fs from 'node:fs';

const manifest = JSON.parse(fs.readFileSync(new URL('../soul-capability-authority.json', import.meta.url), 'utf8'));
const failures = [];

const ownerByCapability = new Map();
for (const [owner, capabilities] of Object.entries(manifest.owners)) {
  for (const id of capabilities) {
    const previous = ownerByCapability.get(id);
    if (previous && previous !== owner) failures.push({ id, reason: 'MULTIPLE_CANONICAL_OWNERS', owners: [previous, owner] });
    ownerByCapability.set(id, owner);
  }
}

const evidence = Object.entries(manifest.evidence ?? {}).flatMap(([owner, items]) =>
  items.map(item => ({ owner, ...item })),
);

for (const item of evidence) {
  const owner = ownerByCapability.get(item.id);
  if (!owner) {
    failures.push({ id: item.id, reason: 'EVIDENCE_WITHOUT_OWNER', evidenceOwner: item.owner });
    continue;
  }
  if (owner !== item.owner) {
    failures.push({ id: item.id, reason: 'EVIDENCE_OWNER_MISMATCH', expected: owner, actual: item.owner });
  }

  const ref = item.ref ?? 'main';
  const rawUrl = 'https://raw.githubusercontent.com/' + item.repository + '/' + ref + '/' + item.path;
  let source = '';
  try {
    const response = await fetch(rawUrl);
    if (!response.ok) throw new Error('HTTP_' + response.status);
    source = await response.text();
  } catch (error) {
    failures.push({ id: item.id, reason: 'EVIDENCE_SOURCE_UNREADABLE', rawUrl, error: error instanceof Error ? error.message : String(error) });
    continue;
  }
  if (!source.includes(item.id)) {
    failures.push({ id: item.id, reason: 'CAPABILITY_ID_NOT_FOUND_IN_EVIDENCE', rawUrl });
  }
}

for (const item of manifest.declaredNotExecutable ?? []) {
  const found = evidence.find(entry => entry.id === item.id);
  if (!found) {
    const rawUrl = 'https://raw.githubusercontent.com/' + item.evidence.repository + '/' + (item.evidence.ref ?? 'main') + '/' + item.evidence.path;
    try {
      const response = await fetch(rawUrl);
      const source = response.ok ? await response.text() : '';
      if (!source.includes(item.id)) failures.push({ id: item.id, reason: 'DECLARED_CAPABILITY_NOT_FOUND', rawUrl });
    } catch (error) {
      failures.push({ id: item.id, reason: 'DECLARED_EVIDENCE_UNREADABLE', rawUrl, error: error instanceof Error ? error.message : String(error) });
    }
  }
}

const affinity = manifest.ossAffinityRouting?.capabilities ?? {};
for (const [id, targets] of Object.entries(affinity)) {
  if (!Array.isArray(targets) || targets.length === 0) {
    failures.push({ id, reason: "AFFINITY_TARGETS_EMPTY" });
  } else if (targets.some(target => !["N01","N02","N03","N04","N05","N06","N07","SARA"].includes(target))) {
    failures.push({ id, reason: "AFFINITY_TARGET_INVALID", targets });
  }
}

const report = {
  status: failures.length === 0 ? 'PASS' : 'FAIL',
  canonicalCapabilities: ownerByCapability.size,
  evidenceEntries: evidence.length,
  declaredNotExecutable: (manifest.declaredNotExecutable ?? []).length,
  failures,
  verifiedAt: new Date().toISOString(),
};
console.log(JSON.stringify(report, null, 2));
if (failures.length) process.exit(1);
