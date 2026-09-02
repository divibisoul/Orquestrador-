import fs from 'node:fs/promises';
import path from 'node:path';

const ROOT = process.cwd();
const MATRIX_PATH = path.join(ROOT, 'soul-nuclei.json');
const SOURCES_ROOT = path.join(ROOT, 'sources');

const DUPLICATION_FAMILIES = [
  {
    id: 'mesh-protocol',
    markers: ['SoulMeshProtocol', 'soul-mesh/1'],
    canonical: 'canonicalMesh',
  },
  {
    id: 'n06-execution',
    markers: ['N06Processor', 'Nucleus06CapabilityRuntime', 'N06CapabilityDispatcher'],
    canonical: 'N06CapabilityEngine',
  },
  {
    id: 'n06-tool-registry',
    markers: ['Nucleus05ToolRegistry', 'N06ToolRegistry'],
    canonical: 'N06ToolRegistry',
  },
];

const PEERS = new Set(['N01', 'N02', 'N03', 'N04', 'N05', 'N06', 'N07']);
const TS_EXTENSIONS = new Set(['.ts', '.tsx', '.js', '.mjs']);
const IGNORED_DIRS = new Set(['node_modules', '.git', '.next', 'dist', 'build', 'coverage']);

async function readJson(filePath) {
  return JSON.parse(await fs.readFile(filePath, 'utf8'));
}

async function walk(dir) {
  const results = [];
  let entries;
  try {
    entries = await fs.readdir(dir, { withFileTypes: true });
  } catch {
    return results;
  }
  for (const entry of entries) {
    if (entry.name.startsWith('.') && entry.name !== '.github') continue;
    if (IGNORED_DIRS.has(entry.name)) continue;
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) results.push(...await walk(fullPath));
    else if (TS_EXTENSIONS.has(path.extname(entry.name))) results.push(fullPath);
  }
  return results;
}

function unique(values) {
  return [...new Set(values)];
}

function addIssue(bucket, issue) {
  bucket.push(issue);
}

const report = {
  system: 'SOUL',
  checker: 'SOUL Architectural Integrity Engine',
  schemaVersion: '1.0.0',
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
if (matrix?.system !== 'SOUL' || matrix?.canonicalMesh?.contractVersion !== '1.1.0') {
  throw new Error('SOUL_MATRIX_INVALID: expected SOUL canonical Mesh contract 1.1.0');
}

report.canonicalAuthorities.push({ responsibility: 'mesh-contract', authority: matrix.canonicalMesh });

for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const peers = Array.isArray(nucleus.peers) ? nucleus.peers : [];
  const invalid = peers.filter((peer) => !PEERS.has(peer) || peer === id);
  if (invalid.length > 0) addIssue(report.topologyConflicts, { nucleus: id, type: 'invalid-peer-set', invalid });
}

for (const [leftId, left] of Object.entries(matrix.nuclei ?? {})) {
  const leftPeers = new Set(left.peers ?? []);
  for (const rightId of leftPeers) {
    const right = matrix.nuclei?.[rightId];
    if (!right) {
      addIssue(report.topologyConflicts, { nucleus: leftId, type: 'peer-not-in-matrix', peer: rightId });
      continue;
    }
    if (!(right.peers ?? []).includes(leftId)) {
      addIssue(report.topologyConflicts, { nucleus: leftId, peer: rightId, type: 'non-symmetric-peer' });
    }
  }
}

const sourceResults = [];
for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const root = path.join(SOURCES_ROOT, id);
  const exists = await fs.stat(root).then((stat) => stat.isDirectory()).catch(() => false);
  if (!exists) {
    addIssue(report.degraded, { nucleus: id, type: 'source-snapshot-unavailable', expected: root });
    continue;
  }
  const files = await walk(root);
  const indexed = [];
  for (const file of files) {
    const content = await fs.readFile(file, 'utf8');
    indexed.push({ file: path.relative(root, file), content });
  }
  sourceResults.push({ id, nucleus, files: indexed });
}

for (const family of DUPLICATION_FAMILIES) {
  const matches = [];
  for (const source of sourceResults) {
    for (const file of source.files) {
      const matchedMarkers = family.markers.filter((marker) => file.content.includes(marker));
      if (matchedMarkers.length > 0) matches.push({ nucleus: source.id, file: file.file, markers: matchedMarkers });
    }
  }
  if (matches.length > 1) {
    const nuclei = unique(matches.map((item) => item.nucleus));
    const files = unique(matches.map((item) => `${item.nucleus}:${item.file}`));
    const issue = { family: family.id, canonical: family.canonical, nuclei, files };
    if (family.id === 'mesh-protocol') addIssue(report.duplicateProtocols, issue);
    else if (family.id === 'n06-tool-registry') addIssue(report.duplicateRegistries, issue);
    else addIssue(report.duplicateAuthorities, issue);
  }
}

for (const source of sourceResults) {
  const protocolFiles = source.files.filter((file) => file.file.includes('SoulMeshProtocol'));
  if (protocolFiles.length > 1) addIssue(report.duplicateProtocols, { nucleus: source.id, type: 'multiple-protocol-files', files: protocolFiles.map((file) => file.file) });
  for (const file of source.files) {
    if (file.content.includes('@deprecated') && (file.file.includes('SoulMeshProtocol') || file.file.includes('ToolRegistry') || file.file.includes('Processor'))) {
      report.legacyAdapters.push({ nucleus: source.id, file: file.file });
    }
    if (file.content.includes('contractVersion') && !file.content.includes('1.1.0')) {
      addIssue(report.contractConflicts, { nucleus: source.id, file: file.file, type: 'non-canonical-contract-version' });
    }
  }
}

for (const [id, nucleus] of Object.entries(matrix.nuclei ?? {})) {
  const declared = Array.isArray(nucleus.capabilities) ? nucleus.capabilities : [];
  if (!declared.length) continue;
  const source = sourceResults.find((item) => item.id === id);
  if (!source) {
    addIssue(report.orphanCapabilities, { nucleus: id, type: 'cannot-verify-without-source', capabilities: declared });
    continue;
  }
  for (const capability of declared) {
    const executable = source.files.some((file) => file.content.includes(`'${capability}'`) || file.content.includes(`\"${capability}\"`));
    if (!executable) addIssue(report.orphanCapabilities, { nucleus: id, capability, type: 'declared-without-executable-evidence' });
  }
}

const critical = [...report.duplicateAuthorities, ...report.duplicateRegistries, ...report.duplicateProtocols, ...report.topologyConflicts, ...report.contractConflicts, ...report.orphanCapabilities];
if (critical.length > 0) report.state = 'FAIL';
else if (report.degraded.length > 0) report.state = 'DEGRADED';

await fs.writeFile(path.join(ROOT, 'SOUL-ARCHITECTURAL-INTEGRITY.json'), `${JSON.stringify(report, null, 2)}\n`, 'utf8');

if (report.state === 'FAIL') process.exitCode = 1;
