import fs from 'node:fs/promises';

const evidence = JSON.parse(await fs.readFile('SOUL-CROSSFRONT-EVIDENCE.json', 'utf8'));
const matrix = JSON.parse(await fs.readFile('soul-nuclei.json', 'utf8'));
const rows = new Map((evidence.rows ?? []).map((row) => [row.id, row]));
const nuclei = matrix.canonicalMesh?.nuclei ?? ['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07'];
const expectedContract = matrix.canonicalMesh?.contractVersion;

if (typeof expectedContract !== 'string' || !expectedContract) {
  throw new Error('CROSSFRONT_CANONICAL_CONTRACT_MISSING');
}

for (const id of nuclei) {
  const row = rows.get(id);
  if (!row) throw new Error(\`CROSSFRONT_NUCLEUS_MISSING:\${id}\`);
  if (row.source === 'not-found' || row.state !== 'OBSERVED') {
    throw new Error(\`CROSSFRONT_EVIDENCE_NOT_OBSERVED:\${id}:\${row.state}:\${row.source}\`);
  }
  if (row.mesh !== true) throw new Error(\`CROSSFRONT_MESH_DISCOVERY_FAILED:\${id}\`);
  if (row.contract !== expectedContract) {
    throw new Error(\`CROSSFRONT_CONTRACT_MISMATCH:\${id}:\${row.contract}:expected:\${expectedContract}\`);
  }
}

console.log(JSON.stringify({
  ok: true,
  expectedContract,
  nuclei: Object.fromEntries(nuclei.map((id) => {
    const row = rows.get(id);
    return [id, { mesh: row.mesh, contract: row.contract, source: row.source, state: row.state }];
  })),
}, null, 2));
