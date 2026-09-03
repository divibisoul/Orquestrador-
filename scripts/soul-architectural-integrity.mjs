import fs from 'node:fs/promises';
import path from 'node:path';

const ROOT = process.cwd();
const MATRIX_PATH = path.join(ROOT, 'soul-nuclei.json');
const SOURCES_ROOT = path.join(ROOT, 'sources');
const API = 'https://api.github.com';

const DUPLICATION_FAMILIES = [
  { id: 'mesh-protocol', definitionPatterns: [/class\s+SoulMeshProtocol\b/, /function\s+createSoulMeshMessage\b/, /export\s+const\s+SOUL_MESH_PROTOCOL\b/], canonical: 'canonicalMesh' },
  { id: 'n06-execution', definitionPatterns: [/class\s+N06Processor\b/, /class\s+Nucleus06CapabilityRuntime\b/, /class\s+N06CapabilityDispatcher\b/, /class\s+N06CapabilityEngine\b/], canonical: 'N06Processor' },
  { id: 'n06-tool-registry', definitionPatterns: [/function\s+createN06Tools\b/, /function\s+createNucleus06Tools\b/, /function\s+createNucleus05Tools\b/, /const\s+N06ToolRegistry\b/], canonical: 'N06ToolRegistry' },
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
  return normalized.includes('/legacy/') || normalized.includes('adapter') || /@deprecated[\s\S]{0,100}compatibility/i.test(content);
}

const report = {
  system: 'SOUL', checker: 'SOUL Architectural Integrity Engine', schemaVersion: '2.0.0', generatedAt: new Date().toISOString(), state: 'PASS',
  duplicateAuthorities: [], duplicateProtocols: [], duplicateRegistries: [], topologyConflicts: [], contractConflicts: [], orphanCapabilities: [],
  legacyAdapters: [], canonicalAuthorities: [], degraded: [], remoteProvenance: [],
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

async function remoteSnapshot(id, nucleus) {
  const token = String(process.env.SOUL_GITHUB_AUDIT_TOKEN ?? '').trim();
  if (!token) {
    report.degraded.push({ nucleus: id, type: 'remote-provenance-unverified', requiredEnv: 'SOUL_GITHUB_AUDIT_TOKEN' });
    return null;
  }
  const headers = { accept: 'application/vnd.github+json', authorization: `Bearer ${token}`, 'x-github-api-version': '2022-11-28', 'user-agent': 'SOUL-Architectural-Integrity-Engine' };
  const repoUrl = `${API}/repos/${nucleus.repository}`;
  const repoResponse = await fetch(repoUrl, { headers });
  if (!repoResponse.ok) throw new Error(`REMOTE_REPO_LOOKUP_FAILED:${id}:${repoResponse.status}`);
  const repo = await repoResponse.json();
  const branch = repo.default_branch;
  const refResponse = await fetch(`${API}/repos/${nucleus.repository}/git/ref/heads/${encodeURIComponent(branch)}`, { headers });
  if (!refResponse.ok) throw new Error(`REMOTE_REF_LOOKUP_FAILED:${id}:${refResponse.status}`);
  const ref = await refResponse.json();
  const actualSha = ref?.object?.sha;
  const expectedSha = nucleus.sourceRef;
  const provenance = { nucleus: id, repository: nucleus.repository, branch, expectedSha, actualSha, state: actualSha === expectedSha ? 'MATCH' : 'DRIFT' };
  report.remoteProvenance.push(provenance);
  if (expectedSha && actualSha !== expectedSha) addIssue(report.contractConflicts, { nucleus: id, type: 'source-ref-drift', expectedSha, actualSha });
  const treeResponse = await fetch(`${API}/repos/${nucleus.repository}/git/trees/${actualSha}?recursive=1`, { headers });
  if (!treeResponse.ok) throw new Error(`REMOTE_TREE_LOOKUP_FAILED:${id}:${treeResponse.status}`);
  const tree = await treeResponse.json();
  if (tree.truncated) addIssue(report.degraded, { nucleus: id, type: 'remote-tree-truncated' });
  const candidates = (tree.tree ?? []).filter((entry) => entry.type === 'blob' && TS_EXTENSIONS.has(path.extname(entry.path)) && entry.size <= 2_000_000);
  const files = [];
  for (const entry of candidates) {
    const blobResponse = await fetch(`${API}/repos/${nucleus.repository}/git/blobs/${entry.sha}`, { headers });
    if (!blobResponse.ok) throw new Error(`REMOTE_BLOB_LOOKUP_FAILED:${id}:${blobResponse.status}:${entry.path}`);
    const blob = await blobResponse.json();
    const content = Buffer.from(String(blob.content ?? ''), 'base64').toString('utf8');
    files.push({ file: entry.path, content });
  }
  return { id, nucleus, files, origin: 'remote', commit: actualSha };
}

const sourceResults = [];
for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const root = path.join(SOURCES_ROOT, id);
  let local = null;
  if (await fs.stat(root).then((stat) => stat.isDirectory()).catch(() => false)) {
    const files = [];
    for (const file of await walk(root)) files.push({ file: path.relative(root, file), content: await fs.readFile(file, 'utf8') });
    local = { id, nucleus, files, origin: 'snapshot', commit: nucleus.sourceRef ?? null };
  }
  let remote = null;
  if (process.env.SOUL_GITHUB_AUDIT_TOKEN) remote = await remoteSnapshot(id, nucleus);
  const selected = remote ?? local;
  if (!selected) {
    if (!report.degraded.some((item) => item.nucleus === id && item.type === 'remote-provenance-unverified')) report.degraded.push({ nucleus: id, type: 'source-snapshot-unavailable' });
    continue;
  }
  sourceResults.push(selected);
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
    if (isCompatibilityAdapter(file.file, file.content) && /SoulMeshProtocol|ToolRegistry|Processor|Runtime/i.test(file.file)) report.legacyAdapters.push({ nucleus: source.id, file: file.file });
    if (file.content.includes('contractVersion') && !file.content.includes('1.1.0')) addIssue(report.contractConflicts, { nucleus: source.id, file: file.file, type: 'non-canonical-contract-version' });
  }
}

const critical = [...report.duplicateAuthorities, ...report.duplicateRegistries, ...report.duplicateProtocols, ...report.topologyConflicts, ...report.contractConflicts, ...report.orphanCapabilities];
if (critical.length) report.state = 'FAIL';
else if (report.degraded.length) report.state = 'DEGRADED';
await fs.writeFile(path.join(ROOT, 'SOUL-ARCHITECTURAL-INTEGRITY.json'), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
if (report.state === 'FAIL' || (process.env.SOUL_REQUIRE_REMOTE_PROVENANCE === 'true' && report.degraded.length)) process.exitCode = 1;
