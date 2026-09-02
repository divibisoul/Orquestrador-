import { createHmac, randomUUID } from 'node:crypto';
import fs from 'node:fs';

// Test-only key. Never use this value for production traffic or secrets.
const TEST_SECRET = 'soul-mesh-contract-test-secret-1.1.0';
const PROTOCOL = 'soul-mesh/1';
const CONTRACT = '1.1.0';
const NUCLEI = ['N01', 'N02', 'N03', 'N04', 'N05', 'N06'];

const pairs = [];
for (let i = 0; i < NUCLEI.length; i += 1) {
  for (let j = i + 1; j < NUCLEI.length; j += 1) {
    pairs.push([NUCLEI[i], NUCLEI[j]]);
  }
}

function requiredEnvelope(source, target) {
  return {
    version: '1.0',
    contractVersion: CONTRACT,
    messageId: randomUUID(),
    source,
    target,
    timestamp: Date.now(),
    nonce: randomUUID(),
    correlationId: `contract-${source}-${target}-${randomUUID()}`,
    type: 'CAPABILITY_REQUEST',
    payload: {
      capability: 'mesh.ping',
      payload: { probe: 'offline-contract' },
    },
  };
}

function canonicalUnsigned(envelope) {
  return JSON.stringify({
    version: envelope.version,
    contractVersion: envelope.contractVersion,
    messageId: envelope.messageId,
    source: envelope.source,
    target: envelope.target,
    timestamp: envelope.timestamp,
    nonce: envelope.nonce,
    correlationId: envelope.correlationId,
    type: envelope.type,
    payload: envelope.payload,
  });
}

function sign(envelope) {
  return createHmac('sha256', TEST_SECRET)
    .update(canonicalUnsigned(envelope), 'utf8')
    .digest('hex');
}

function assertEnvelope(envelope, signature) {
  if (envelope.version !== '1.0') throw new Error('version mismatch');
  if (envelope.contractVersion !== CONTRACT) throw new Error('contract mismatch');
  if (!envelope.messageId) throw new Error('messageId missing');
  if (!NUCLEI.includes(envelope.source) || !NUCLEI.includes(envelope.target)) throw new Error('unknown nucleus');
  if (envelope.source === envelope.target) throw new Error('source/target identical');
  if (!Number.isFinite(envelope.timestamp) || envelope.timestamp <= 0) throw new Error('invalid timestamp');
  if (!envelope.nonce) throw new Error('nonce missing');
  if (!envelope.correlationId) throw new Error('correlationId missing');
  if (envelope.type !== 'CAPABILITY_REQUEST') throw new Error('type mismatch');
  if (envelope.payload?.capability !== 'mesh.ping') throw new Error('capability missing');
  if (envelope.payload?.payload === undefined) throw new Error('nested payload missing');

  const expected = sign(envelope);
  if (expected !== signature) throw new Error('HMAC mismatch');
}

const results = [];
for (const [a, b] of pairs) {
  for (const [source, target] of [[a, b], [b, a]]) {
    try {
      const envelope = requiredEnvelope(source, target);
      const signature = sign(envelope);
      const serialized = JSON.stringify({ ...envelope, hmac: signature });
      const roundTrip = JSON.parse(serialized);
      assertEnvelope(roundTrip, roundTrip.hmac);
      results.push({ source, target, status: 'PASS', correlationId: roundTrip.correlationId, contractVersion: CONTRACT });
    } catch (error) {
      results.push({ source, target, status: 'FAIL', error: error instanceof Error ? error.message : String(error) });
    }
  }
}

const failed = results.filter((item) => item.status !== 'PASS');
const report = {
  system: 'SOUL',
  mode: 'offline-contract-prover',
  protocol: PROTOCOL,
  contractVersion: CONTRACT,
  pairCount: pairs.length,
  directedEnvelopeCount: results.length,
  overallStatus: failed.length === 0 ? 'PASS' : 'FAIL',
  results,
  note: 'No network requests are performed. The embedded HMAC key exists only for deterministic contract validation.',
};

fs.mkdirSync('.diagnostics', { recursive: true });
fs.writeFileSync('.diagnostics/contract-15-pairs-valid.json', `${JSON.stringify(report, null, 2)}\n`, 'utf8');
console.log(JSON.stringify(report, null, 2));

if (failed.length > 0) process.exitCode = 2;
