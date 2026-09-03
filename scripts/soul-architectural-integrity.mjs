import fs from 'node:fs/promises';
import path from 'node:path';

const ROOT = process.cwd();
const MATRIX_PATH = path.join(ROOT, 'soul-nuclei.json');
const SOURCES_ROOT = path.join(ROOT, 'sources');

const DUPLICATION_FAMILIES = [
  {
    id: 'mesh-protocol',
    definitionPatterns: [/class\s+SoulMeshProtocol\b/, /function\s+createSoulMeshMessage\b/, /export\s+const\s+SOUL_MESH_PROTOCOL\b/],
    canonical: 'canonicalMesh',
  },
  {
    id: 'n06-execution',
    definitionPatterns: [/class\s+N06Processor\b/, /class\s+Nucleus06CapabilityRuntime\b/, /class\s+N06CapabilityDispatcher\b/, /class\s+N06CapabilityEngine\b/],
    canonical: 'N06CapabilityEngine',
  },
  {
    id: 'n06-tool-registry',
    definitionPatterns: [/function\s+createN06Tools\b/, /function\s+createNucleus06Tools\b/, /function\s+createNucleus05Tools\b/, /const\s+N06ToolRegistry\b/],
    canonical: 'N06ToolRegistry',
  },
];

const PEERS = new Set(['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07']);
const TS_EXTENSIONS = new Set(['.ts', '.tsx', '.js', '.mjs']);
const IGNORED_DIRS = new Set(['node_modules', '.git', '.next', 'dist', 'build', 'coverage']);

async function readJson(filePath) { return JSON.parse(await fs.readFile(filePath, 'utf8')); }
async function walk(dir) {
  const results = [];
  let entries;
  try { entries = await fs.readdir(dir, { withFileTypes: true }); } catch { return results; }
  for (const entry of entries) {
    if (entry.name.startsWith('.') && entry.name !== '.github') continue;
    if (IGNORED_DIRS.has(entry.name)) continue;
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) results.push(...await walk(fullPath));
    else if (TS_EXTENSIONS.has(path.extname(entry.name))) results.push(fullPath);
  }
  return results;
}
function addIssue(bucket, issue) { bucket.push(issue); }
function isCompatibilityAdapter(relativePath, content) {
  const normalized = relativePath.replaceAll('\\', '/').toLowerCase();
  return normalized.includes('/legacy/') || normalized.includes('adapter') || /@deprecated\s+.*compatibility/i.test(content);
}

const report = {
  system: 'SOUL',
  checker: 'SOUL Architectural Integrity Engine',
  schemaVersion: '1.0.1',
  generatedAt: new Date().toISOString(),
  state: 'PASS',
  duplicateAuthorities: [],
  duplicateProtocols: [],
  duplicateRegistries: [],
  topologyConflicts: [],
  contractConflicts: [],
  orphanCapabilities: [],
  legacyAdapters: [],
  canonicalAuthorities: [],
  degraded: [],
};

const matrix = await readJson(MATRIX_PATH);
if (matrix?.system !== 'SOUL' || matrix?.canonicalMesh?.contractVersion !== '1.1.0') throw new Error('SOUL_MATRIX_INVALID: expected SOUL canonical Mesh contract 1.1.0');
report.canonicalAuthorities.push({ responsibility: 'mesh-contract', authority: matrix.canonicalMesh });

for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const peers = Array.isArray(nucleus.peers) ? nucleus.peers : [];
  const invalid = peers.filter((peer) => !PEERS.has(peer) || peer === id);
  if (invalid.length) addIssue(report.topologyConflicts, { nucleus: id, type: 'invalid-peer-set', invalid });
}
for (const [leftId, left] of Object.entries(matrix.nuclei ?? {})) {
  for (const rightId of new Set(left.peers ?? [])) {
    const right = matrix.nuclei?.[rightId];
    if (!right) addIssue(report.topologyConflicts, { nucleus: leftId, type: 'peer-not-in-matrix', peer: rightId });
    else if (!(right.peers ?? []).includes(leftId)) addIssue(report.topologyConflicts, { nucleus: leftId, peer: rightId, type: 'non-symmetric-peer' });
  }
}

const sourceResults = [];
for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const root = path.join(SOURCES_ROOT, id);
  if (!(await fs.stat(root).then((stat) => stat.isDirectory()).catch(() => false))) {
    addIssue(report.degraded, { nucleus: id, type: 'source-snapshot-unavailable', expected: root });
    continue;
  }
  const files = [];
  for (const file of await walk(root)) files.push({ file: path.relative(root, file), content: await fs.readFile(file, 'utf8') });
  sourceResults.push({ id, nucleus, files });
}

for (const family of DUPLICATION_FAMILIES) {
  const authorities = [];
  for (const source of sourceResults) {
    for (const file of source.files) {
      if (isCompatibilityAdapter(file.file, file.content)) {
        if (family.definitionPatterns.some((pattern) => pattern.test(file.content))) report.legacyAdapters.push({ nucleus: source.id, family: family.id, file: file.file });
        continue;
      }
      const definitions = family.definitionPatterns.filter((pattern) => pattern.test(file.content)).map(String);
      if (definitions.length) authorities.push({ nucleus: source.id, file: file.file, definitions });
    }
  }
  if (authorities.length > 1) {
    const issue = { family: family.id, canonical: family.canonical, authorities };
    if (family.id === 'mesh-protocol') addIssue(report.duplicateProtocols, issue);
    else if (family.id === 'n06-tool-registry') addIssue(report.duplicateRegistries, issue);
    else addIssue(report.duplicateAuthorities, issue);
  }
}

for (const source of sourceResults) {
  for (const file of source.files) {
    if (isCompatibilityAdapter(file.file, file.content) && (file.file.includes('SoulMeshProtocol') || file.file.includes('ToolRegistry') || file.file.includes('Processor') || file.file.includes('Runtime'))) report.legacyAdapters.push({ nucleus: source.id, file: file.file });
    if (file.content.includes('contractVersion') && !file.content.includes('1.1.0')) addIssue(report.contractConflicts, { nucleus: source.id, file: file.file, type: 'non-canonical-contract-version' });
  }
}

const critical = [...report.duplicateAuthorities, ...report.duplicateRegistries, ...report.duplicateProtocols, ...report.topologyConflicts, ...report.contractConflicts, ...report.orphanCapabilities];
if (critical.length) report.state = 'FAIL';
else if (report.degraded.length) report.state = 'DEGRADED';
await fs.writeFile(path.join(ROOT, 'SOUL-ARCHITECTURAL-INTEGRITY.json'), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
if (report.state === 'FAIL') process.exitCode = 1;
