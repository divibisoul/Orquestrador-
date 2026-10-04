import { createHmac, randomUUID, timingSafeEqual } from 'node:crypto';
import { readFileSync } from 'node:fs';

const SECRET = process.env.SOUL_MESH_HMAC_SECRET ?? '';
const TOKEN = process.env.SOUL_MESH_TOKEN ?? '';
const N04_URL = (process.env.N04_URL ?? 'http://127.0.0.1:3000').replace(/\/$/, '');
const N05_URL = (process.env.N05_URL ?? 'http://127.0.0.1:3001').replace(/\/$/, '');

function envelope(source, target, capability, payload, correlationId, id = randomUUID()) {
  const nonce = randomUUID().replaceAll('-', '').slice(0, 32);
  const message = { protocol:'soul-mesh/1', contractVersion:'1.1.0', id, correlationId, source, target, kind:'request', capability, payload, timestamp:Date.now(), meta:{runtime:'stage-02-harness',transport:'HTTP',encoding:'json',version:'1.1.0',traceId:correlationId} };
  return { message, nonce, hmac: sign(message, nonce) };
}
function sign(message, nonce) {
  const canonical = JSON.stringify({protocol:message.protocol,contractVersion:message.contractVersion,id:message.id,correlationId:message.correlationId,source:message.source,target:message.target,kind:message.kind,capability:message.capability ?? null,payload:message.payload,timestamp:message.timestamp,meta:message.meta ?? null,nonce});
  return createHmac('sha256', SECRET).update(canonical,'utf8').digest('hex');
}
function verifyResponse(request, body) {
  const nonce = body.nonce ?? body.meta?.nonce ?? '';
  const hmac = body.hmac ?? '';
  if (!nonce || !hmac) throw new Error('STAGE02_RESPONSE_HMAC_MISSING');
  if (body.source !== request.target || body.target !== request.source) throw new Error('STAGE02_ROUTE_MISMATCH');
  if (body.correlationId !== request.correlationId) throw new Error('STAGE02_CORRELATION_MISMATCH');
  const expected = (() => { const canonical = JSON.stringify({protocol:body.protocol,contractVersion:body.contractVersion,id:body.id,correlationId:body.correlationId,source:body.source,target:body.target,kind:body.kind,capability:body.capability ?? null,payload:body.payload,timestamp:body.timestamp,meta:body.meta ?? null,nonce}); return createHmac('sha256',SECRET).update(canonical,'utf8').digest('hex'); })();
  const a=Buffer.from(hmac,'hex'),b=Buffer.from(expected,'hex');
  if(a.length!==b.length || !timingSafeEqual(a,b)) throw new Error('STAGE02_RESPONSE_HMAC_INVALID');
}
async function send(url, source, target, capability, payload, correlationId, opts={}) {
  const {message,nonce,hmac}=envelope(source,target,capability,payload,correlationId);
  const controller=new AbortController();
  if(opts.timeoutMs) setTimeout(()=>controller.abort(),opts.timeoutMs).unref();
  const headers={'content-type':'application/json','x-soul-mesh-nonce':nonce,'x-soul-mesh-hmac':hmac, ...(opts.bearer===false?{}:{authorization:'Bearer '+TOKEN})};
  const response=await fetch(url+'/api/soul-mesh',{method:'POST',headers,body:JSON.stringify({...message,nonce}),signal:controller.signal});
  const body=await response.json().catch(()=>null);
  return {response,body,message};
}

async function expectNative() {
  const correlationId='stage02-root-correlation';
  const a=await send(N04_URL,'N05','N04','context-orchestration',{stage:'N05_N04',seed:true},correlationId,{bearer:false});
  if(!a.response.ok) throw new Error('N05_TO_N04_HTTP_'+a.response.status);
  verifyResponse(a.message,a.body);
  if(a.body.kind!=='response') throw new Error('N05_TO_N04_NOT_RESPONSE');
  console.log(JSON.stringify({state:'REAL',direction:'N05->N04',capability:a.message.capability,correlationId,responseMessageId:a.body.id}));
  const b=await send(N05_URL,'N04','N05','mesh.resident.describe@1.0.0',{stage:'N05_N04',upstream:a.body.payload},correlationId,{bearer:true});
  if(!b.response.ok) throw new Error('N04_TO_N05_HTTP_'+b.response.status);
  verifyResponse(b.message,b.body);
  if(b.body.kind!=='response') throw new Error('N04_TO_N05_NOT_RESPONSE');
  console.log(JSON.stringify({state:'REAL',direction:'N04->N05',capability:b.message.capability,correlationId,responseMessageId:b.body.id}));
  return {a,b,correlationId};
}

async function authzDenial() {
  const correlationId='stage02-authz-denial';
  const x=await send(N04_URL,'N05','N04','tool.run',{tool:'getWeather',arguments:{latitude:0,longitude:0}},correlationId,{bearer:false});
  if(x.response.status!==401) throw new Error('EXPECTED_N04_SESSION_BOUNDARY_401_GOT_'+x.response.status);
  console.log(JSON.stringify({state:'REAL',test:'tool.run-without-user-session',expectedStatus:401,actualStatus:x.response.status}));
}

async function payloadLimit() {
  const correlationId='stage02-payload-limit';
  const big='x'.repeat(2*1024*1024);
  const {message,nonce,hmac}=envelope('N05','N04','context-orchestration',{blob:big},correlationId);
  const response=await fetch(N04_URL+'/api/soul-mesh',{method:'POST',headers:{'content-type':'application/json','x-soul-mesh-nonce':nonce,'x-soul-mesh-hmac':hmac},body:JSON.stringify({...message,nonce})});
  if(response.status!==413) throw new Error('EXPECTED_413_GOT_'+response.status);
  console.log(JSON.stringify({state:'REAL',test:'payload-limit',expectedStatus:413,actualStatus:response.status}));
}

async function retrySemantics() {
  const correlationId = 'stage02-retry-correlation';
  let transientFailure = false;
  try {
    await send(N04_URL, 'N05', 'N04', 'context-orchestration', { attempt: 1, idempotencyKey: 'stage02-retry-1' }, correlationId, { bearer: false, timeoutMs: 1 });
  } catch {
    transientFailure = true;
  }
  const retry = await send(N04_URL, 'N05', 'N04', 'context-orchestration', { attempt: 2, idempotencyKey: 'stage02-retry-1' }, correlationId, { bearer: false });
  if (!retry.response.ok) throw new Error('RETRY_FINAL_ATTEMPT_HTTP_' + retry.response.status);
  verifyResponse(retry.message, retry.body);
  if (retry.body.kind !== 'response') throw new Error('RETRY_FINAL_ATTEMPT_NOT_RESPONSE');
  console.log(JSON.stringify({
    state: 'REAL',
    test: 'idempotent-retry',
    transientFailureObserved: transientFailure,
    correlationId,
    retryMessageId: retry.body.id
  }));
}

async function main() {
  await expectNative();
  await authzDenial();
  await payloadLimit();
  await retrySemantics();
  console.log(JSON.stringify({state:'REAL',stage:'N05<->N04',composition:'PASS',provenance:{source:'runtime-evidence',correlationId:'stage02-root-correlation'}}));
}

main().catch(error=>{console.error(error instanceof Error?error.stack:String(error));process.exit(1);});