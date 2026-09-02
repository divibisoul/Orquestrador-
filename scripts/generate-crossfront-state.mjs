import fs from 'node:fs/promises';
import path from 'node:path';

const ROOT = process.cwd();
const nuclei = [
  ['N01','aeternum-core-29',['src/core/soul-mesh/peerClient.ts','build.gradle.kts','app/build.gradle.kts']],
  ['N02','Eternium-',['api/soul-mesh.ts','src/soul-mesh/SoulMeshProtocol.ts','package.json']],
  ['N03','nexus-aeternum-fusion',['api/soul-mesh.ts','src/soul-mesh/SoulMeshProtocol.ts','package.json']],
  ['N04','nextjs-ai-chatbots',['lib/soul-mesh/SoulMeshCapabilities.ts','lib/soul-mesh/SoulMeshProtocol.ts','package.json']],
  ['N05','nextjs-ai-chatbot',['src/soul-mesh/SoulMeshAdapter.ts','src/soul-mesh/SoulMeshProtocol.ts','package.json']],
  ['N06','nextjs-ai-chatbot-2000',['app/api/soul-mesh/route.ts','src/soul-mesh/SoulMeshProtocol.ts','package.json']],
  ['N07','Orquestrador-',['N07_SYSTEM_STATE.md','go.mod','mesh/mesh.go']],
];

async function readFirst(root, candidates) {
  for (const relative of candidates) {
    try { const content = await fs.readFile(path.join(root, relative), 'utf8'); return { path: relative, content }; } catch {}
  }
  return null;
}
function contractOf(text) { return text.match(/(?:contractVersion|SOUL_MESH_CONTRACT_VERSION|contract)\s*[:=]\s*["'`]?(1\.\d+\.\d+)/i)?.[1] ?? 'not-detected'; }
function has(text, terms) { return terms.some(term => text.toLowerCase().includes(term)); }

const rows = [];
for (const [id, repo, candidates] of nuclei) {
  const root = path.join(ROOT, 'sources', id);
  const primary = await readFirst(root, candidates);
  const all = [];
  for (const candidate of candidates) { try { all.push([candidate, await fs.readFile(path.join(root,candidate),'utf8')]); } catch {} }
  const joined = all.map(([,text]) => text).join('\n');
  rows.push({ id, repo, contract: contractOf(joined), mesh: has(joined,['soul-mesh/1','soul mesh']), neuralBridge: has(joined,['synapticnodebridge','neural','federation']), source: primary?.path ?? 'not-found' });
}

const lines = [
  '# SOUL — Crossfront State (Generated)',
  '',
  '> Generated automatically by N07 governance tooling. This file reports repository evidence; it does not fabricate runtime availability.',
  '',
  '| Nucleus | Repository | Contract detected | Mesh evidence | Neural/federation evidence | Source |',
  '|---|---|---:|---|---|---|',
  ...rows.map(r => `| ${r.id} | ${r.repo} | ${r.contract} | ${r.mesh ? 'present' : 'not detected'} | ${r.neuralBridge ? 'present' : 'not detected'} | ${r.source} |`),
  '',
  '## Governance rules',
  '',
  '- Canonical contract target: **1.1.0**.',
  '- N07 remains the orchestration, federation and SuperGPU control plane.',
  '- A missing external deployment URL is reported as uncommissioned, never as healthy.',
  '- Repository evidence is separated from live runtime evidence.',
  '- This report is generated from the checked-out repository state and is intended to be consumed by the N07 release/commissioning gates.',
  '',
];
await fs.writeFile(path.join(ROOT,'SOUL-SYSTEM-CROSSFRONT-STATE.md'), lines.join('\n'), 'utf8');
