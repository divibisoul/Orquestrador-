import fs from 'node:fs/promises';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const ROOT = process.cwd();
const MATRIX_PATH = path.join(ROOT, 'soul-nuclei.json');
const MASTER_CONTRACT_PATH = path.join(ROOT, 'SOUL_MASTER_ENGINEERING_CONTRACT.md');
const SOURCES_ROOT = path.join(ROOT, 'sources');
const API = 'https://api.github.com';
const IDS = ['N01','N02','N03','N04','N05','N06','N07'];
const PEERS = new Set(IDS);
const EXPECTED_ADJACENCY = {
  N01:['N02'], N02:['N01','N03'], N03:['N02','N04'], N04:['N03','N05'],
  N05:['N04','N06'], N06:['N05','N07'], N07:['N06'],
};
const TS_EXTENSIONS = new Set(['.ts','.tsx','.js','.mjs']);
const IGNORED_DIRS = new Set(['node_modules','.git','.next','dist','build','coverage']);
const DUPLICATION_FAMILIES = [
  { id:'mesh-protocol', patterns:[/class\s+SoulMeshProtocol\b/,/function\s+createSoulMeshMessage\b/,/export\s+const\s+SOUL_MESH_PROTOCOL\b/], canonical:'canonicalMesh' },
  { id:'n06-execution', patterns:[/class\s+N06Processor\b/,/class\s+Nucleus06CapabilityRuntime\b/,/class\s+N06CapabilityDispatcher\b/,/class\s+N06CapabilityEngine\b/], canonical:'N06CapabilityEngine' },
  { id:'n06-tool-registry', patterns:[/function\s+createN06Tools\b/,/function\s+createNucleus06Tools\b/,/const\s+N06ToolRegistry\b/], canonical:'N06ToolRegistry' },
];

async function readJson(file) { return JSON.parse(await fs.readFile(file,'utf8')); }
async function walk(dir) {
  const out=[]; let entries=[];
  try { entries=await fs.readdir(dir,{withFileTypes:true}); } catch { return out; }
  for (const e of entries) {
    if (IGNORED_DIRS.has(e.name)) continue;
    if (e.name.startsWith('.') && e.name !== '.github') continue;
    const full=path.join(dir,e.name);
    if (e.isDirectory()) out.push(...await walk(full));
    else if (TS_EXTENSIONS.has(path.extname(e.name))) out.push(full);
  }
  return out;
}
function add(bucket, issue) { bucket.push(issue); }
function isAdapter(file, content='') {
  const p=file.replaceAll('\\','/').toLowerCase();
  return p.includes('/legacy/') || p.includes('adapter') || /@deprecated[\s\S]{0,120}compatibility/i.test(content);
}

const report={system:'SOUL',checker:'SOUL Architectural Integrity Engine',schemaVersion:'3.0.0',generatedAt:new Date().toISOString(),state:'PASS',duplicateAuthorities:[],duplicateProtocols:[],protocolImplementations:[],duplicateRegistries:[],topologyConflicts:[],contractConflicts:[],orphanCapabilities:[],legacyAdapters:[],canonicalAuthorities:[],degraded:[],remoteProvenance:[],candidateProvenance:[],promptContract:[],headSnapshot:[],canonicalMeshContracts:[],nonCanonicalProtocolCopies:[],protocolImplementations:[]};
const matrix=await readJson(MATRIX_PATH);

if (matrix?.system!=='SOUL') add(report.contractConflicts,{type:'system-invalid'});
if (matrix?.canonicalMesh?.protocol!=='soul-mesh/1' || matrix?.canonicalMesh?.contractVersion!=='1.1.0') add(report.contractConflicts,{type:'canonical-mesh-invalid',expected:'soul-mesh/1@1.1.0'});
if (!Array.isArray(matrix?.canonicalMesh?.nuclei) || matrix.canonicalMesh.nuclei.length!==7 || matrix.canonicalMesh.nuclei.some((id,i)=>id!==IDS[i])) add(report.topologyConflicts,{type:'seven-nucleus-order-invalid'});
if (matrix?.fusion?.topology!=='linear-adjacent' || matrix?.fusion?.policy!=='adjacent-only-dynamic' || matrix?.fusion?.bidirectionalChannelsPerEdge!==2) add(report.topologyConflicts,{type:'fusion-policy-invalid'});
if (matrix?.hardware?.controlPlane!=='N07' || matrix?.hardware?.realHardwareOnly!==true || matrix?.hardware?.simulation!==false) add(report.contractConflicts,{type:'hardware-policy-invalid'});
if (!(await fs.stat(MASTER_CONTRACT_PATH).then(s=>s.isFile()).catch(()=>false))) add(report.promptContract,{type:'master-contract-missing',path:'SOUL_MASTER_ENGINEERING_CONTRACT.md'});
else report.promptContract.push({type:'master-contract-present'});

for (const id of IDS) if (!matrix.nuclei?.[id]) add(report.topologyConflicts,{nucleus:id,type:'nucleus-missing'});
for (const id of IDS) {
  const n=matrix.nuclei?.[id] ?? {};
  const peers=Array.isArray(n.peers)?n.peers:[];
  const invalid=peers.filter(p=>!PEERS.has(p)||p===id);
  if (invalid.length) add(report.topologyConflicts,{nucleus:id,type:'invalid-peer-set',invalid});
  const expected=EXPECTED_ADJACENCY[id] ?? [];
  const actual=Array.isArray(n.adjacentPeers)?[...n.adjacentPeers].sort():[];
  if (JSON.stringify(actual)!==JSON.stringify([...expected].sort())) add(report.topologyConflicts,{nucleus:id,type:'adjacency-mismatch',expected,actual});
  if (id!=='N07' && (!/^[0-9a-f]{40}$/.test(String(n.sourceRef??'')))) add(report.contractConflicts,{nucleus:id,type:'source-ref-invalid',sourceRef:n.sourceRef});
}
for (const [id,n] of Object.entries(matrix.nuclei??{})) for (const peer of n.peers??[]) {
  if (!matrix.nuclei?.[peer]) add(report.topologyConflicts,{nucleus:id,type:'peer-not-in-matrix',peer});
  else if (!(matrix.nuclei[peer].peers??[]).includes(id)) add(report.topologyConflicts,{nucleus:id,peer,type:'non-symmetric-peer'});
}

async function remoteSnapshot(id,nucleus) {
  const token=String(process.env.SOUL_GITHUB_AUDIT_TOKEN || process.env.GITHUB_TOKEN || '').trim();
  if (!token) { add(report.degraded,{nucleus:id,type:'remote-provenance-unverified',requiredEnv:'SOUL_GITHUB_AUDIT_TOKEN'}); return null; }
  const headers={accept:'application/vnd.github+json',authorization:`Bearer ${token}`,'x-github-api-version':'2022-11-28','user-agent':'SOUL-Architectural-Integrity-Engine'};
  const readRemote=async(url)=>{const response=await fetch(url,{headers});if(!response.ok)throw new Error('REMOTE_API_FAILED:'+id+':'+response.status+':'+url);return response.json();};
  const repo=await readRemote(`${API}/repos/${nucleus.repository}`);
  const branch=repo.default_branch;
  const ref=await readRemote(`${API}/repos/${nucleus.repository}/git/ref/heads/${encodeURIComponent(branch)}`);
  const actualSha=ref?.object?.sha;
  if (!actualSha) throw new Error('REMOTE_HEAD_SHA_MISSING:'+id);
  const expectedSha=String(nucleus.sourceRef??'');
  report.remoteProvenance.push({nucleus:id,repository:nucleus.repository,branch,expectedSha,actualSha,state:actualSha===expectedSha?'MATCH':'DRIFT'});
  if (id!=='N07' && /^[0-9a-f]{40}$/.test(expectedSha) && actualSha!==expectedSha) add(report.contractConflicts,{nucleus:id,type:'source-ref-drift',expectedSha,actualSha});
  let snapshotSha=/^[0-9a-f]{40}$/.test(expectedSha)?expectedSha:actualSha;
  const validationPullRequest=Number(nucleus.validationPullRequest??0);
  if (validationPullRequest > 0) {
    if (!Number.isInteger(validationPullRequest)) throw new Error('VALIDATION_PULL_REQUEST_INVALID:'+id);
    const candidate=await readRemote(`${API}/repos/${nucleus.repository}/pulls/${validationPullRequest}`);
    const candidateSha=String(candidate?.head?.sha??'');
    const candidateRepo=String(candidate?.head?.repo?.full_name??'');
    if (!/^[0-9a-f]{40}$/.test(candidateSha) || candidateRepo!==nucleus.repository) {
      throw new Error('VALIDATION_PULL_REQUEST_HEAD_INVALID:'+id+':'+validationPullRequest);
    }
    if (candidate.state!=='open' && candidate.merged!==true) {
      throw new Error('VALIDATION_PULL_REQUEST_NOT_ACTIVE:'+id+':'+validationPullRequest);
    }
    report.candidateProvenance.push({
      nucleus:id,
      repository:nucleus.repository,
      pullRequest:validationPullRequest,
      headRef:String(candidate?.head?.ref??''),
      candidateSha,
      mainBranch:branch,
      expectedMainSha:expectedSha,
      actualMainSha:actualSha,
      draft:Boolean(candidate?.draft),
      state:candidate?.merged===true?'MERGED_CANDIDATE':'OPEN_PR_CANDIDATE',
    });
    snapshotSha=candidateSha;
  }
  const tree=await readRemote(`${API}/repos/${nucleus.repository}/git/trees/${snapshotSha}?recursive=1`);
  if (tree.truncated) throw new Error('REMOTE_TREE_TRUNCATED:'+id+':'+snapshotSha);
  const evidencePaths=new Set([...(nucleus.entrypoints??[]),...(nucleus.contractEvidencePaths??[])]);
  const relevantPath=/(?:SoulMeshProtocol|soul-mesh|N06|N06ToolRegistry|createN06Tools|createNucleus06Tools)/i;
  const candidates=(tree.tree??[]).filter(entry=>entry.type==='blob'&&TS_EXTENSIONS.has(path.extname(entry.path))&&(evidencePaths.has(entry.path)||relevantPath.test(entry.path)));
  if (candidates.length===0) throw new Error('REMOTE_SNAPSHOT_EMPTY:'+id+':'+snapshotSha);
  if (candidates.length>180) throw new Error('REMOTE_SNAPSHOT_CANDIDATE_LIMIT_EXCEEDED:'+id+':'+candidates.length);
  const files=[];
  for (const entry of candidates) {
    if (!entry.sha) throw new Error('REMOTE_BLOB_SHA_MISSING:'+id+':'+entry.path);
    const blob=await readRemote(`${API}/repos/${nucleus.repository}/git/blobs/${entry.sha}`);
    if (blob.encoding!=='base64'||typeof blob.content!=='string') throw new Error('REMOTE_BLOB_CONTENT_INVALID:'+id+':'+entry.path);
    files.push({file:entry.path,content:Buffer.from(blob.content,'base64').toString('utf8')});
  }
  if (!files.length) throw new Error('REMOTE_SNAPSHOT_EMPTY:'+id+':'+snapshotSha);
  return {id,nucleus,origin:'remote',commit:snapshotSha,files};
}

async function workspaceSnapshot(nucleus) {
  const commit=execFileSync('git',['rev-parse','HEAD'],{cwd:ROOT,encoding:'utf8'}).trim();
  const tracked=execFileSync('git',['ls-files','-z'],{cwd:ROOT,encoding:'utf8'}).split('\0').filter(Boolean);
  const files=[];
  for (const file of tracked) {
    if (!TS_EXTENSIONS.has(path.extname(file))) continue;
    files.push({file,content:await fs.readFile(path.join(ROOT,file),'utf8')});
  }
  if (!files.length) throw new Error('WORKSPACE_SNAPSHOT_EMPTY:N07:'+commit);
  return {id:'N07',nucleus,origin:'working-tree',commit,files};
}

const sourceResults=[];
for (const [id,nucleus] of Object.entries(matrix.nuclei??{})) {
  const root=path.join(SOURCES_ROOT,id); let local=null;
  if (id==='N07') {
    try { local=await workspaceSnapshot(nucleus); }
    catch (error) { add(report.degraded,{nucleus:id,type:'workspace-snapshot-unavailable',error:String(error)}); }
  } else if (await fs.stat(root).then(s=>s.isDirectory()).catch(()=>false)) {
    const files=[]; for (const f of await walk(root)) files.push({file:path.relative(root,f),content:await fs.readFile(f,'utf8')});
    local={id,nucleus,origin:'snapshot',commit:nucleus.sourceRef??null,files};
  }
  let remote=null;
  if (id!=='N07') {
    try { remote=await remoteSnapshot(id,nucleus); }
    catch (error) { add(report.degraded,{nucleus:id,type:'remote-snapshot-unavailable',error:String(error)}); }
  }
  const selected=id==='N07'?local:(remote??local);
  if (!selected) add(report.degraded,{nucleus:id,type:'source-snapshot-unavailable'});
  else if (!Array.isArray(selected.files)||selected.files.length===0) add(report.degraded,{nucleus:id,type:'source-snapshot-empty',origin:selected.origin??'unknown',commit:selected.commit??null});
  else sourceResults.push(selected);
}

for (const family of DUPLICATION_FAMILIES) {
  // A shared wire protocol is implemented at each nucleus boundary. Identical
  // protocol symbols in separate repositories are implementations, not multiple
  // global authorities. N07's declarative canonicalMesh remains the one authority.
  if (family.id==='mesh-protocol') continue;

  const authorities=[];
  for (const source of sourceResults) for (const file of source.files) {
    if (isAdapter(file.file,file.content)) {
      if (family.patterns.some(p=>p.test(file.content))) report.legacyAdapters.push({nucleus:source.id,family:family.id,file:file.file});
      continue;
    }
    const defs=family.patterns.filter(p=>p.test(file.content)).map(String);
    if (defs.length) authorities.push({nucleus:source.id,file:file.file,definitions:defs});
  }
  if (authorities.length>1) {
    const issue={family:family.id,canonical:family.canonical,authorities};
    if (family.id==='n06-tool-registry') add(report.duplicateRegistries,issue);
    else add(report.duplicateAuthorities,issue);
  }
}

// The cross-nucleus protocol is one wire contract with a canonical implementation
// per independent runtime. Validate each declared canonical module; do not mistake
// identical symbol names in separate repositories for duplicate global authorities.
const protocolFamily = DUPLICATION_FAMILIES.find(f => f.id === 'mesh-protocol');
const protocolImplementations = [];
const canonicalMeshContracts = [];
const nonCanonicalProtocolCopies = [];
const normalizeRepoPath = value => String(value ?? '').replaceAll('\\', '/').replace(/^\.\/+/, '');
const expectedMeshKinds = ['request', 'response', 'event', 'error'];
for (const source of sourceResults) {
  // N07's canonical Mesh envelope is Go-owned; the TypeScript contract gate applies to N01-N06.
  if (source.id === 'N07') continue;

  const canonicalPath = normalizeRepoPath(source.nucleus.canonicalMeshContractPath);
  if (!canonicalPath) {
    add(report.contractConflicts, { nucleus: source.id, type: 'canonical-mesh-contract-path-missing' });
    continue;
  }

  const canonicalFile = source.files.find(file => normalizeRepoPath(file.file) === canonicalPath);
  if (!canonicalFile) {
    add(report.contractConflicts, {
      nucleus: source.id,
      type: 'canonical-mesh-contract-file-missing',
      path: canonicalPath,
      commit: source.commit ?? null,
    });
    canonicalMeshContracts.push({ nucleus: source.id, file: canonicalPath, state: 'MISSING' });
    continue;
  }

  const content = canonicalFile.content;
  const protocolMatch = content.match(/(?:export\s+)?const\s+SOUL_MESH_PROTOCOL\s*=\s*['"]([^'"]+)['"]/);
  const versionMatch = content.match(/(?:export\s+)?const\s+SOUL_MESH_CONTRACT_VERSION\s*=\s*['"]([^'"]+)['"]/);
  const kindMatch = content.match(/\btype\s+SoulMeshKind\s*=\s*([^;\n]+)/);
  const kinds = kindMatch ? Array.from(kindMatch[1].matchAll(/['"]([^'"]+)['"]/g), match => match[1]) : [];
  const nucleiMatch = content.match(/(?:export\s+)?const\s+SOUL_NUCLEI\b[^=]*=\s*\[([^\]]+)\]/);
  const nuclei = nucleiMatch ? Array.from(nucleiMatch[1].matchAll(/['"](N0[1-7])['"]/g), match => match[1]) : [];
  const checks = {
    protocol: protocolMatch?.[1] === matrix.canonicalMesh.protocol,
    contractVersion: versionMatch?.[1] === matrix.canonicalMesh.contractVersion,
    canonicalKinds: JSON.stringify([...kinds].sort()) === JSON.stringify([...expectedMeshKinds].sort()),
    numericTimestamp: /\btimestamp\s*:\s*number\b/.test(content),
    completeNucleusSet: JSON.stringify(nuclei) === JSON.stringify(IDS),
    messageFactory: /\bfunction\s+createSoulMeshMessage\b/.test(content),
    messageValidator: /\b(?:function|const)\s+(?:validateSoulMeshMessage|validateMessage|isSoulMeshMessage)\b/.test(content),
    messageEnvelopeFields: /\bid\s*:\s*string\b/.test(content)
      && /\bcorrelationId\s*:\s*string\b/.test(content)
      && /\bsource\s*:/.test(content)
      && /\btarget\s*:/.test(content)
      && /\bkind\s*:/.test(content)
      && /\bpayload\s*:/.test(content)
      && /\btimestamp\s*:/.test(content),
  };
  const failedChecks = Object.entries(checks).filter(([, passed]) => !passed).map(([name]) => name);
  canonicalMeshContracts.push({
    nucleus: source.id,
    file: canonicalPath,
    protocol: protocolMatch?.[1] ?? null,
    contractVersion: versionMatch?.[1] ?? null,
    kinds,
    nuclei,
    checks,
    state: failedChecks.length ? 'FAIL' : 'PASS',
  });
  if (failedChecks.length) {
    add(report.contractConflicts, {
      nucleus: source.id,
      file: canonicalPath,
      type: 'canonical-mesh-contract-invalid',
      failedChecks,
      expectedProtocol: matrix.canonicalMesh.protocol,
      expectedContractVersion: matrix.canonicalMesh.contractVersion,
      expectedKinds: expectedMeshKinds,
      expectedNuclei: IDS,
    });
  }

  const activePaths = new Set([
    ...(source.nucleus.entrypoints ?? []),
    ...(source.nucleus.contractEvidencePaths ?? []),
  ].map(normalizeRepoPath));

  for (const file of source.files) {
    const relativePath = normalizeRepoPath(file.file);
    const definitions = protocolFamily.patterns.filter(pattern => pattern.test(file.content)).map(String);
    if (!definitions.length) continue;

    if (isAdapter(file.file, file.content)) {
      report.legacyAdapters.push({ nucleus: source.id, family: protocolFamily.id, file: relativePath });
      continue;
    }

    const fileProtocol = file.content.match(/(?:export\s+)?const\s+SOUL_MESH_PROTOCOL\s*=\s*['"]([^'"]+)['"]/);
    const fileVersion = file.content.match(/(?:export\s+)?const\s+SOUL_MESH_CONTRACT_VERSION\s*=\s*['"]([^'"]+)['"]/);
    const isCanonical = relativePath === canonicalPath;
    const isActiveBinding = activePaths.has(relativePath);
    const role = isCanonical ? 'canonical-contract' : (isActiveBinding ? 'active-runtime-binding' : 'non-canonical-copy');
    protocolImplementations.push({
      nucleus: source.id,
      file: relativePath,
      role,
      protocol: fileProtocol?.[1] ?? null,
      contractVersion: fileVersion?.[1] ?? null,
      definitions,
    });

    if (isCanonical) continue;

    if (isActiveBinding) {
      // A declared active route must import/re-export its canonical schema, not define another one.
      add(report.duplicateProtocols, {
        family: protocolFamily.id,
        canonical: canonicalPath,
        authorities: [{
          nucleus: source.id,
          file: relativePath,
          definitions,
          reason: 'active entrypoint declares Mesh protocol symbols outside its canonical module',
        }],
      });
      continue;
    }

    // Keep old or auxiliary copies visible for later adapter reconciliation; their presence
    // alone is not evidence of a second active authority in the declared runtime path.
    nonCanonicalProtocolCopies.push({
      nucleus: source.id,
      file: relativePath,
      canonical: canonicalPath,
      protocol: fileProtocol?.[1] ?? null,
      contractVersion: fileVersion?.[1] ?? null,
      containsLegacyAck: /\b(?:kind\s*:\s*[^;\n]*|type\s+\w*MeshKind\s*=)\s*[^;\n]*['"]ack['"]/.test(file.content),
      containsStringTimestamp: /\btimestamp\s*:\s*string\b/.test(file.content),
      state: 'NON_CANONICAL_COPY_RETAINED',
    });
  }
}
report.canonicalMeshContracts = canonicalMeshContracts;
report.nonCanonicalProtocolCopies = nonCanonicalProtocolCopies;
report.protocolImplementations = protocolImplementations;

const n06Source = sourceResults.find(source => source.id === 'N06');
const n06EngineFiles = n06Source
  ? n06Source.files.filter(f=>/N06CapabilityEngine\.(ts|tsx|js|mjs)$/.test(f.file) && !isAdapter(f.file,f.content))
  : [];
if (n06Source && !n06EngineFiles.length) add(report.contractConflicts,{nucleus:'N06',type:'canonical-capability-engine-not-found'});

const n03Source = sourceResults.find(source => source.id === 'N03');
const n03CanonicalPath = normalizeRepoPath(n03Source?.nucleus.canonicalMeshContractPath);
const n03ActivePaths = new Set([
  ...(n03Source?.nucleus.entrypoints ?? []),
  ...(n03Source?.nucleus.contractEvidencePaths ?? []),
].map(normalizeRepoPath));
const n03Legacy = n03Source
  ? n03Source.files.filter(f =>
      /(^|\/)soul-mesh\/SoulMeshProtocol\.(ts|tsx|js|mjs)$/.test(normalizeRepoPath(f.file))
      && normalizeRepoPath(f.file) !== n03CanonicalPath
      && n03ActivePaths.has(normalizeRepoPath(f.file))
      && !isAdapter(f.file, f.content)
      && protocolFamily.patterns.some(p => p.test(f.content)))
  : [];
if (n03Legacy.length) add(report.duplicateProtocols, {
  family: 'mesh-protocol',
  canonical: n03CanonicalPath,
  authorities: n03Legacy.map(f => ({ nucleus: 'N03', file: f.file, reason: 'active N03 entrypoint declares a second protocol schema' })),
});

const critical=[...report.duplicateAuthorities,...report.duplicateRegistries,...report.duplicateProtocols,...report.topologyConflicts,...report.contractConflicts,...report.orphanCapabilities];
if (critical.length) report.state='FAIL'; else if (report.degraded.length) report.state='DEGRADED';
await fs.writeFile(path.join(ROOT,'SOUL-ARCHITECTURAL-INTEGRITY.json'),`${JSON.stringify(report,null,2)}\n`,'utf8');
console.log(JSON.stringify(report,null,2));
if (report.state==='FAIL' || (process.env.SOUL_REQUIRE_REMOTE_PROVENANCE==='true' && report.degraded.length)) process.exitCode=1;
