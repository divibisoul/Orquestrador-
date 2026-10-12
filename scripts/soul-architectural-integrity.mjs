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

const report={system:'SOUL',checker:'SOUL Architectural Integrity Engine',schemaVersion:'3.0.0',generatedAt:new Date().toISOString(),state:'PASS',duplicateAuthorities:[],duplicateProtocols:[],protocolImplementations:[],duplicateRegistries:[],topologyConflicts:[],contractConflicts:[],orphanCapabilities:[],legacyAdapters:[],canonicalAuthorities:[],degraded:[],remoteProvenance:[],promptContract:[],headSnapshot:[]};
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
  const snapshotSha=/^[0-9a-f]{40}$/.test(expectedSha)?expectedSha:actualSha;
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

// Record distributed Mesh implementations separately and validate their declared
// wire values against the canonical contract instead of calling every peer
// implementation a second global authority.
const protocolFamily=DUPLICATION_FAMILIES.find(f=>f.id==='mesh-protocol');
const protocolImplementations=[];
for (const source of sourceResults) {
  const declaredPaths=new Set(source.nucleus.contractEvidencePaths??[]);
  for (const file of source.files) {
    if (!protocolFamily.patterns.some(p=>p.test(file.content))) continue;
    if (isAdapter(file.file,file.content)) {
      report.legacyAdapters.push({nucleus:source.id,family:protocolFamily.id,file:file.file});
      continue;
    }
    const protocolMatch=file.content.match(/(?:export\s+)?const\s+SOUL_MESH_PROTOCOL\s*=\s*['"]([^'"]+)['"]/);
    const versionMatch=file.content.match(/(?:export\s+)?const\s+SOUL_MESH_CONTRACT_VERSION\s*=\s*['"]([^'"]+)['"]/);
    protocolImplementations.push({
      nucleus:source.id,
      file:file.file,
      role:declaredPaths.has(file.file)?'declared-contract-evidence':'peer-implementation',
      protocol:protocolMatch?.[1]??null,
      contractVersion:versionMatch?.[1]??null,
    });
    if (protocolMatch && protocolMatch[1]!==matrix.canonicalMesh.protocol) {
      add(report.contractConflicts,{nucleus:source.id,file:file.file,type:'mesh-protocol-version-mismatch',expected:matrix.canonicalMesh.protocol,actual:protocolMatch[1]});
    }
    if (versionMatch && versionMatch[1]!==matrix.canonicalMesh.contractVersion) {
      add(report.contractConflicts,{nucleus:source.id,file:file.file,type:'mesh-contract-version-mismatch',expected:matrix.canonicalMesh.contractVersion,actual:versionMatch[1]});
    }
    if (/(?:interface|type)\s+SoulMeshMessage\b/.test(file.content) && !/\bcontractVersion\b/.test(file.content)) {
      add(report.contractConflicts,{nucleus:source.id,file:file.file,type:'mesh-contract-version-missing',expected:matrix.canonicalMesh.contractVersion});
    }
  }
}
report.protocolImplementations=protocolImplementations;

const n06Source = sourceResults.find(source => source.id === 'N06');
const n06EngineFiles = n06Source
  ? n06Source.files.filter(f=>/N06CapabilityEngine\.(ts|tsx|js|mjs)$/.test(f.file) && !isAdapter(f.file,f.content))
  : [];
if (n06Source && !n06EngineFiles.length) add(report.contractConflicts,{nucleus:'N06',type:'canonical-capability-engine-not-found'});

const n03Source = sourceResults.find(source => source.id === 'N03');
const n03Legacy = n03Source
  ? n03Source.files.filter(f=>/(^|\/)soul-mesh\/SoulMeshProtocol\.(ts|tsx|js|mjs)$/.test(f.file) && !isAdapter(f.file,f.content))
  : [];
if (n03Legacy.length) add(report.duplicateProtocols,{family:'mesh-protocol',canonical:'lib/soul-mesh/SoulMeshProtocol.ts',authorities:n03Legacy.map(f=>({nucleus:'N03',file:f.file,reason:'legacy protocol is active, not adapter'}))});

const critical=[...report.duplicateAuthorities,...report.duplicateRegistries,...report.duplicateProtocols,...report.topologyConflicts,...report.contractConflicts,...report.orphanCapabilities];
if (critical.length) report.state='FAIL'; else if (report.degraded.length) report.state='DEGRADED';
await fs.writeFile(path.join(ROOT,'SOUL-ARCHITECTURAL-INTEGRITY.json'),`${JSON.stringify(report,null,2)}\n`,'utf8');
console.log(JSON.stringify(report,null,2));
if (report.state==='FAIL' || (process.env.SOUL_REQUIRE_REMOTE_PROVENANCE==='true' && report.degraded.length)) process.exitCode=1;
