import fs from 'node:fs/promises';
import path from 'node:path';

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

const token=String(process.env.SOUL_GITHUB_AUDIT_TOKEN??'').trim();
const requireRemote=process.env.SOUL_REQUIRE_REMOTE_PROVENANCE==='true';
const report={system:'SOUL',checker:'SOUL Architectural Integrity Engine',schemaVersion:'3.1.0',generatedAt:new Date().toISOString(),state:'PASS',provenanceMode:token?'REMOTE_REQUIRED':'LOCAL_STRUCTURAL_DEGRADED',duplicateAuthorities:[],duplicateProtocols:[],duplicateRegistries:[],topologyConflicts:[],contractConflicts:[],orphanCapabilities:[],legacyAdapters:[],canonicalAuthorities:[],degraded:[],remoteProvenance:[],promptContract:[],headSnapshot:[]};
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
  if (!token) { add(report.degraded,{nucleus:id,type:'remote-provenance-unverified',requiredEnv:'SOUL_GITHUB_AUDIT_TOKEN',fallback:'LOCAL_STRUCTURAL'}); return null; }
  const headers={accept:'application/vnd.github+json',authorization:`Bearer ${token}`,'x-github-api-version':'2022-11-28','user-agent':'SOUL-Architectural-Integrity-Engine'};
  const repoResponse=await fetch(`${API}/repos/${nucleus.repository}`,{headers});
  if (!repoResponse.ok) throw new Error(`REMOTE_REPO_LOOKUP_FAILED:${id}:${repoResponse.status}`);
  const repo=await repoResponse.json(); const branch=repo.default_branch;
  const refResponse=await fetch(`${API}/repos/${nucleus.repository}/git/ref/heads/${encodeURIComponent(branch)}`,{headers});
  if (!refResponse.ok) throw new Error(`REMOTE_REF_LOOKUP_FAILED:${id}:${refResponse.status}`);
  const ref=await refResponse.json(); const actualSha=ref?.object?.sha; const expectedSha=nucleus.sourceRef;
  const provenance={nucleus:id,repository:nucleus.repository,branch,expectedSha,actualSha,state:actualSha===expectedSha?'MATCH':'DRIFT'}; report.remoteProvenance.push(provenance);
  if (id!=='N07' && actualSha!==expectedSha) add(report.contractConflicts,{nucleus:id,type:'source-ref-drift',expectedSha,actualSha});
  return {id,nucleus,origin:'remote',commit:actualSha,files:[]};
}

const sourceResults=[];
for (const [id,nucleus] of Object.entries(matrix.nuclei??{})) {
  const root=path.join(SOURCES_ROOT,id); let local=null;
  if (await fs.stat(root).then(s=>s.isDirectory()).catch(()=>false)) {
    const files=[]; for (const f of await walk(root)) files.push({file:path.relative(root,f),content:await fs.readFile(f,'utf8')});
    local={id,nucleus,origin:'snapshot',commit:nucleus.sourceRef??null,files};
  }
  const remote=token ? await remoteSnapshot(id,nucleus) : null;
  const selected=remote ?? local;
  if (!selected) add(report.degraded,{nucleus:id,type:'source-snapshot-unavailable'});
  else sourceResults.push(selected);
}

for (const family of DUPLICATION_FAMILIES) {
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
    if (family.id==='mesh-protocol') add(report.duplicateProtocols,issue);
    else if (family.id==='n06-tool-registry') add(report.duplicateRegistries,issue);
    else add(report.duplicateAuthorities,issue);
  }
}

const n06EngineFiles=sourceResults.flatMap(s=>s.files.filter(f=>s.id==='N06' && /N06CapabilityEngine\.(ts|tsx|js|mjs)$/.test(f.file) && !isAdapter(f.file,f.content)));
if (!n06EngineFiles.length) add(report.contractConflicts,{nucleus:'N06',type:'canonical-capability-engine-not-found'});

const n03Legacy=sourceResults.flatMap(s=>s.files.filter(f=>s.id==='N03' && /(^|\/)soul-mesh\/SoulMeshProtocol\.(ts|tsx|js|mjs)$/.test(f.file) && !isAdapter(f.file,f.content)));
if (n03Legacy.length) add(report.duplicateProtocols,{family:'mesh-protocol',canonical:'src/mesh/SoulMeshProtocol.ts',authorities:n03Legacy.map(f=>({nucleus:'N03',file:f.file,reason:'legacy protocol is active, not adapter'}))});

const critical=[...report.duplicateAuthorities,...report.duplicateRegistries,...report.duplicateProtocols,...report.topologyConflicts,...report.contractConflicts,...report.orphanCapabilities];
if (critical.length) report.state='FAIL'; else if (report.degraded.length) report.state='DEGRADED';
await fs.writeFile(path.join(ROOT,'SOUL-ARCHITECTURAL-INTEGRITY.json'),`${JSON.stringify(report,null,2)}\n`,'utf8');
if (report.state==='FAIL' || (requireRemote && token && report.degraded.length)) process.exitCode=1;
