import fs from 'node:fs/promises';
import path from 'node:path';

const ROOT = process.cwd();
const MATRIX_PATH = path.join(ROOT, 'soul-nuclei.json');
const matrix = JSON.parse(await fs.readFile(MATRIX_PATH, 'utf8'));

async function readOptional(root, relative, missing) {
  try {
    return { path: relative, content: await fs.readFile(path.join(root, relative), 'utf8') };
  } catch (error) {
    missing.push({ path: relative, reason: error instanceof Error ? error.message : String(error) });
    return null;
  }
}

function contractOf(text) {
  // Prefer explicit canonical constants over arbitrary object fields: telemetry
  // and nested legacy payloads may carry an unrelated contractVersion.
  const canonicalPatterns = [
    /\bSOUL_MESH_CONTRACT_VERSION\s*=\s*["'\x60]((?:1)\.\d+\.\d+)["'\x60]/,
    /\bSoulMeshContractVersion\s*=\s*["'\x60]((?:1)\.\d+\.\d+)["'\x60]/,
    /\bCONTRACT_VERSION\s*=\s*["'\x60]((?:1)\.\d+\.\d+)["'\x60]/,
    /\bContractVersion\s*[:=]\s*["'\x60]((?:1)\.\d+\.\d+)["'\x60]/,
  ];
  for (const pattern of canonicalPatterns) {
    const match = text.match(pattern);
    if (match) return match[1];
  }
  return text.match(/(?:contractVersion|contract)\s*[:=]\s*["'\x60]?((?:1)\.\d+\.\d+)/i)?.[1] ?? 'not-detected';
}
function has(text, terms) { const normalized = text.toLowerCase(); return terms.some((term) => normalized.includes(String(term).toLowerCase())); }

const rows = [];
for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const missing = [];
  const root = id === 'N07' ? ROOT : path.join(ROOT, 'sources', id);
  const all = [];
  for (const candidate of nucleus.entrypoints ?? []) {
    const item = await readOptional(root, candidate, missing);
    if (item) all.push(item);
  }
  const contractEvidence = [];
  for (const candidate of nucleus.contractEvidencePaths ?? []) {
    try { contractEvidence.push(await fs.readFile(path.join(root, candidate), 'utf8')); } catch { /* optional evidence path; missing does not degrade a valid entrypoint */ }
  }
  const joined = [...all.map((item) => item.content), ...contractEvidence].join('\n');
  const source = all[0]?.path ?? 'not-found';
  const mesh = has(joined, ['soul-mesh/1', 'soul mesh', 'SOUL_MESH_PROTOCOL', 'SOUL_MESH_VERSION']);
  rows.push({
    id,
    repo: nucleus.repository,
    role: nucleus.role,
    contract: contractOf(joined),
    mesh,
    neuralBridge: has(joined, ['synapticnodebridge', 'neural', 'federation']),
    source,
    // File presence alone is not Mesh evidence; fail closed on missing Mesh proof.
    state: missing.length > 0 || source === 'not-found' || !mesh ? 'DEGRADED' : 'OBSERVED',
    missing,
  });
}

const lines = [
  '# SOUL — Crossfront State (Generated)',
  '',
  '> Generated from the declarative SOUL matrix. Missing repository evidence is surfaced as DEGRADED; runtime availability is never fabricated.',
  '',
  '| Nucleus | Repository | Role | Contract | Mesh | Neural/Federation | Evidence state | Source |',
  '|---|---|---|---:|---|---|---|---|',
  ...rows.map((row) => `| ${row.id} | ${row.repo} | ${row.role} | ${row.contract} | ${row.mesh ? 'present' : 'not detected'} | ${row.neuralBridge ? 'present' : 'not detected'} | ${row.state} | ${row.source} |`),
  '',
  '## Governance',
  '',
  `- Canonical Mesh contract target: **${matrix.canonicalMesh?.protocol ?? 'soul-mesh/1'} / ${matrix.canonicalMesh?.contractVersion ?? 'not-declared'}**.`,
  '- SOUL is one system; N01–N07 are specialized nuclei.',
  '- The matrix is the source of truth for nucleus identity, role and topology.',
  '- A legacy implementation may remain only as a compatibility adapter pointing at a canonical authority.',
  '- Repository evidence and live runtime evidence are kept separate.',
  '',
];

await fs.writeFile(path.join(ROOT, 'SOUL-SYSTEM-CROSSFRONT-STATE.md'), lines.join('\n'), 'utf8');
await fs.writeFile(path.join(ROOT, 'SOUL-CROSSFRONT-EVIDENCE.json'), `${JSON.stringify({ system: 'SOUL', generatedAt: new Date().toISOString(), rows }, null, 2)}\n`, 'utf8');
