import fs from 'node:fs/promises';

const evidence = JSON.parse(await fs.readFile('SOUL-CROSSFRONT-EVIDENCE.json', 'utf8'));
const rows = new Map((evidence.rows ?? []).map((row) => [row.id, row]));

for (const id of ['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07']) {
  if (!rows.has(id)) throw new Error(`CROSSFRONT_NUCLEUS_MISSING:${id}`);
}

const n05 = rows.get('N05');
if (!n05 || n05.mesh !== true || n05.source === 'not-found' || n05.state !== 'OBSERVED') {
  throw new Error('CROSSFRONT_N05_MESH_DISCOVERY_FAILED');
}

const n07 = rows.get('N07');
if (!n07 || n07.mesh !== true || n07.source === 'not-found' || n07.state !== 'OBSERVED') {
  throw new Error('CROSSFRONT_N07_MESH_DISCOVERY_FAILED');
}

if (n05.contract === 'not-detected') throw new Error('CROSSFRONT_N05_CONTRACT_DISCOVERY_FAILED');
if (n07.contract === 'not-detected') throw new Error('CROSSFRONT_N07_CONTRACT_DISCOVERY_FAILED');

console.log(JSON.stringify({
  ok: true,
  N05: { mesh: n05.mesh, contract: n05.contract, source: n05.source, state: n05.state },
  N07: { mesh: n07.mesh, contract: n07.contract, source: n07.source, state: n07.state },
}, null, 2));
