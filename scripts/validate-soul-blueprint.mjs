import fs from "node:fs/promises";
import process from "node:process";

const required = [
  "ARCHITECTURE.md","INFRASTRUCTURE.md","MESH_CONTRACT.md","CAPABILITY_REGISTRY.md",
  "OBSERVABILITY.md","DEPLOYMENT.md","LOCAL_DEV.md","E2E_TEST_PLAN.md","RUNBOOK.md",
  "contracts/soul-mesh/envelope.schema.json",
  "contracts/soul-mesh/capability-registration.schema.json",
  "contracts/soul-mesh/agent-registration.schema.json",
  "contracts/soul-mesh/mesh-combo.schema.json",
  "contracts/soul-mesh/hop-result.schema.json",
  "contracts/soul-mesh/sara-cycle.schema.json",
  "contracts/soul-mesh/sara-trace.schema.json",
  "contracts/soul-mesh/handoff.schema.json",
  "contracts/soul-config.schema.json",
  "contracts/soul-mesh/openapi.yaml","config/soul-config.v1.yaml","infra/local/docker-compose.yml"
];

const failures=[];
for(const f of required){try{await fs.access(f)}catch{failures.push("missing:"+f)}}
const env=JSON.parse(await fs.readFile("contracts/soul-mesh/envelope.schema.json","utf8"));
for(const k of ["version","contractVersion","messageId","source","target","timestamp","nonce","correlationId","type","payload"]){
  if(!env.required.includes(k)) failures.push("envelope-required:"+k);
}
if(env.properties.contractVersion.const!=="1.1.0") failures.push("contract-version");
if(env["x-soul-max-payload-bytes"]!==2097152) failures.push("payload-limit");
if(env["x-soul-correlation-rule"]!=="preserve across hops") failures.push("correlation-rule");

const cfg=await fs.readFile("config/soul-config.v1.yaml","utf8");
for(const needle of [
  'contractVersion: "1.1.0"','timeoutMs: 30000','maxPayloadBytes: 2097152',
  'failureThreshold: 5','openSeconds: 60','globalWorkers: 32',
  'meshHmacRequired: true','mtlsRequired: true'
]) if(!cfg.includes(needle)) failures.push("config:"+needle);

if(failures.length){
  console.error(JSON.stringify({state:"BLOCKED",failures},null,2));
  process.exit(1);
}
console.log(JSON.stringify({state:"VALIDATED",requiredFiles:required.length},null,2));
