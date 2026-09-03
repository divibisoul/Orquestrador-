import fs from 'node:fs/promises';

const API='https://api.github.com';
const IDS=['N01','N02','N03','N04','N05','N06','N07'];
const REPOS={N01:'divibisoul/aeternum-core-29',N02:'divibisoul/Eternium-',N03:'divibisoul/nexus-aeternum-fusion',N04:'divibisoul/nextjs-ai-chatbots',N05:'divibisoul/nextjs-ai-chatbot',N06:'divibisoul/nextjs-ai-chatbot-2000',N07:'divibisoul/Orquestrador-'};
const token=String(process.env.SOUL_GITHUB_AUDIT_TOKEN??'').trim();
const headers={accept:'application/vnd.github+json','x-github-api-version':'2022-11-28','user-agent':'SOUL-Live-Source-Audit',...(token?{authorization:`Bearer ${token}`}:{})};
const report={system:'SOUL',checker:'SOUL Live Source Audit',schemaVersion:'1.0.0',generatedAt:new Date().toISOString(),state:'PASS',nuclei:[],failures:[],degraded:[]};
const read=async(url)=>{const r=await fetch(url,{headers});if(!r.ok)throw new Error(`HTTP_${r.status}:${url}`);return r.json();};
for(const id of IDS){
  const repo=REPOS[id];
  try{
    const meta=await read(`${API}/repos/${repo}`);
    const branch=meta.default_branch;
    const ref=await read(`${API}/repos/${repo}/git/ref/heads/${encodeURIComponent(branch)}`);
    const sha=ref.object.sha;
    const item={id,repository:repo,branch,head:sha,checkedFiles:[],checks:{}};
    const tree=await read(`${API}/repos/${repo}/git/trees/${sha}?recursive=1`);
    const paths=(tree.tree??[]).filter(x=>x.type==='blob').map(x=>x.path);
    const candidates=paths.filter(p=>/SoulMeshProtocol\.(ts|tsx|js|mjs)$|N06(CapabilityEngine|Processor|CapabilityRuntime)\.(ts|tsx|js|mjs)$|N06ToolRegistry|soul-mesh\.ts$|PrefrontalNeocortex|neocortex\.(ts|tsx|go)$|fusion|supergpu/i.test(p));
    for(const p of candidates.slice(0,120)){
      const entry=(tree.tree??[]).find(x=>x.path===p); if(!entry?.sha)continue;
      const blob=await read(`${API}/repos/${repo}/git/blobs/${entry.sha}`);
      const content=Buffer.from(String(blob.content??''),'base64').toString('utf8');
      item.checkedFiles.push(p);
      if(id==='N03' && /(^|\/)src\/soul-mesh\/SoulMeshProtocol\.(ts|tsx|js|mjs)$/.test(p) && !/deprecated|compatibility|adapter/i.test(content)) report.failures.push({nucleus:id,type:'active-legacy-mesh-protocol',file:p});
      if(id==='N06' && /N06Processor\.(ts|tsx|js|mjs)$/.test(p) && !/deprecated|facade|compatibility/i.test(content)) report.failures.push({nucleus:id,type:'non-adapter-n06-processor-authority',file:p});
      if(id==='N06' && /Nucleus06CapabilityRuntime\.(ts|tsx|js|mjs)$/.test(p) && !/facade|compatibility/i.test(content)) report.failures.push({nucleus:id,type:'non-adapter-n06-runtime-authority',file:p});
    }
    item.checks.headResolved=true; item.checks.treeResolved=true; report.nuclei.push(item);
  }catch(error){report.degraded.push({nucleus:id,type:'live-source-audit-failed',error:String(error)});}
}
if(report.failures.length)report.state='FAIL';else if(report.degraded.length)report.state='DEGRADED';
await fs.writeFile('SOUL-LIVE-SOURCE-AUDIT.json',`${JSON.stringify(report,null,2)}\n`,'utf8');
console.log(JSON.stringify(report,null,2));
if(report.state!=='PASS')process.exitCode=1;
