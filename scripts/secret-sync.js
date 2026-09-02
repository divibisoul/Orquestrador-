import { mkdir, writeFile } from 'node:fs/promises';
import sodium from 'libsodium-wrappers';

const apiBase = (process.env.GITHUB_API_URL || 'https://api.github.com').replace(/\/$/, '');
const token = process.env.REPO_SECRETS_ADMIN_TOKEN?.trim();
const secretValue = process.env.SOUL_MESH_TOKEN;
const secretName = process.env.SYNC_SECRET_NAME?.trim() || 'SOUL_MESH_TOKEN';
const repositories = (process.env.TARGET_REPOSITORIES || '')
  .split(',')
  .map((value) => value.trim())
  .filter(Boolean);

if (!token) throw new Error('REPO_SECRETS_ADMIN_TOKEN is required');
if (!secretValue) throw new Error('SOUL_MESH_TOKEN is required');
if (repositories.length === 0) throw new Error('TARGET_REPOSITORIES is empty');

await sodium.ready;

const headers = {
  Accept: 'application/vnd.github+json',
  Authorization: `Bearer ${token}`,
  'X-GitHub-Api-Version': '2026-03-10',
  'User-Agent': 'SOUL-N07-SecretSync',
};

async function github(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, { ...options, headers: { ...headers, ...(options.headers || {}) } });
  const text = await response.text();
  let body = null;
  try { body = text ? JSON.parse(text) : null; } catch { body = { raw: text.slice(0, 500) }; }
  if (!response.ok) {
    const detail = body?.message || body?.error || `HTTP ${response.status}`;
    throw new Error(`${options.method || 'GET'} ${path}: ${detail}`);
  }
  return body;
}

const report = {
  generatedAt: new Date().toISOString(),
  secretName,
  repositories: [],
};

for (const repository of repositories) {
  const result = { repository, status: 'failed' };
  try {
    const publicKey = await github(`/repos/${repository}/actions/secrets/public-key`);
    if (!publicKey?.key_id || !publicKey?.key) throw new Error('Repository public key response is incomplete');

    const plaintext = sodium.from_string(secretValue);
    const keyBytes = sodium.from_base64(publicKey.key, sodium.base64_variants.ORIGINAL);
    const encrypted = sodium.crypto_box_seal(plaintext, keyBytes);
    const encryptedValue = sodium.to_base64(encrypted, sodium.base64_variants.ORIGINAL);

    await github(`/repos/${repository}/actions/secrets/${encodeURIComponent(secretName)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ encrypted_value: encryptedValue, key_id: publicKey.key_id }),
    });

    const metadata = await github(`/repos/${repository}/actions/secrets/${encodeURIComponent(secretName)}`);
    if (metadata?.name !== secretName) throw new Error('Secret metadata verification failed');

    result.status = 'synced';
    result.verified = true;
    result.updatedAt = metadata.updated_at || null;
  } catch (error) {
    result.error = error instanceof Error ? error.message : String(error);
  }
  report.repositories.push(result);
  console.log(JSON.stringify({ repository, status: result.status, verified: result.verified === true, error: result.error || null }));
}

await mkdir('.diagnostics', { recursive: true });
await writeFile('.diagnostics/secret-sync-report.json', JSON.stringify(report, null, 2));

const failed = report.repositories.filter((item) => item.status !== 'synced');
if (failed.length > 0) {
  throw new Error(`Secret synchronization failed for ${failed.length} of ${report.repositories.length} repositories`);
}
