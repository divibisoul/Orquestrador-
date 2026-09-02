import fs from 'node:fs';

const nuclei = [
  { id: 'N01', env: 'VITE_SOUL_MESH_N07_URL', secret: 'VITE_SOUL_MESH_TOKEN' },
  { id: 'N02', env: 'SOUL_MESH_N07_URL', secret: 'SOUL_MESH_HMAC_SECRET' },
  { id: 'N03', env: 'SOUL_MESH_N07_URL', secret: 'SOUL_MESH_HMAC_SECRET' },
  { id: 'N04', env: 'SOUL_MESH_N07_URL', secret: 'SOUL_MESH_HMAC_SECRET' },
  { id: 'N05', env: 'SOUL_N07_URL', secret: 'SOUL_MESH_HMAC_SECRET' },
  { id: 'N06', env: 'SOUL_MESH_N07_URL', secret: 'SOUL_MESH_HMAC_SECRET' },
];

const urlPattern = /^https?:\/\/[^\s/]+(?:\/[^\s]*)?$/i;
const state = {
  generatedAt: new Date().toISOString(),
  contractVersion: '1.1.0',
  protocol: 'soul-mesh/1',
  mode: 'static-only',
  nuclei: {},
};

for (const nucleus of nuclei) {
  const url = (process.env[nucleus.env] ?? '').trim();
  const validUrl = url === '' || urlPattern.test(url);
  state.nuclei[nucleus.id] = {
    requiredEndpointVariable: nucleus.env,
    endpointConfigured: url !== '',
    endpointFormatValid: validUrl,
    secretVariable: nucleus.secret,
    readiness: url && validUrl ? 'READY_FOR_CONNECTION' : url ? 'INVALID_CONFIGURATION' : 'NOT_CONFIGURED',
  };
}

state.summary = {
  ready: Object.values(state.nuclei).filter((item) => item.readiness === 'READY_FOR_CONNECTION').length,
  notConfigured: Object.values(state.nuclei).filter((item) => item.readiness === 'NOT_CONFIGURED').length,
  invalid: Object.values(state.nuclei).filter((item) => item.readiness === 'INVALID_CONFIGURATION').length,
  note: 'This validator never performs network requests and never reads or prints secret values. Presence of a secret is intentionally not inferred from process environment unless the variable itself is supplied to the runner.',
};

fs.mkdirSync('.diagnostics', { recursive: true });
fs.writeFileSync('.diagnostics/topology-readiness.json', `${JSON.stringify(state, null, 2)}\n`, 'utf8');
console.log(JSON.stringify(state, null, 2));
