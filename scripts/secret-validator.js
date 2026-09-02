import { mkdir, writeFile } from 'node:fs/promises';

const apiBase = (process.env.GITHUB_API_URL || 'https://api.github.com').replace(/\/$/, '');
const token = process.env.REPO_SECRETS_ADMIN_TOKEN?.trim();
const secretName = process.env.SYNC_SECRET_NAME?.trim() || 'SOUL_MESH_TOKEN';
const repositories = (process.env.TARGET_REPOSITORIES || '')
  .split(',')
  .map((value) => value.trim())
  .filter(Boolean);
const maxAgeHours = Number(process.env.MAX_SECRET_AGE_HOURS || 36);

if (!token) throw new Error('REPO_SECRETS_ADMIN_TOKEN is required');
if (!Number.isFinite(maxAgeHours) || maxAgeHours <= 0) throw new Error('MAX_SECRET_AGE_HOURS must be positive');
if (repositories.length === 0) throw new Error('TARGET_REPOSITORIES is empty');

const headers = {
  Accept: 'application/vnd.github+json',
  Authorization: `Bearer ${token}`,
  'X-GitHub-Api-Version': '2026-03-10',
  'User-Agent': 'SOUL-N07-SecretValidator',
};

async function github(path) {
  const response = await fetch(`${apiBase}${path}`, { headers });
  const text = await response.text();
  let body = null;
  try { body = text ? JSON.parse(text) : null; } catch { body = { raw: text.slice(0, 500) }; }
  if (!response.ok) throw new Error(`${path}: ${body?.message || `HTTP ${response.status}`}`);
  return body;
}

const now = Date.now();
const report = {
  generatedAt: new Date(now).toISOString(),
  secretName,
  verificationMode: 'metadata-presence-and-freshness',
  note: 'GitHub never exposes plaintext secret values through the API; this validator proves presence and freshness, not plaintext equality.',
  maxAgeHours,
  lastSuccessfulSync: null,
  resyncSuggested: false,
  repositories: [],
};

for (const repository of repositories) {
  const result = { repository, status: 'missing', present: false, fresh: false, cureSuggestion: null };
  try {
    const metadata = await github(`/repos/${repository}/actions/secrets/${encodeURIComponent(secretName)}`);
    result.present = metadata?.name === secretName;
    result.updatedAt = metadata?.updated_at || null;

    if (result.updatedAt) {
      const updatedAtMs = Date.parse(result.updatedAt);
      const ageHours = (now - updatedAtMs) / 36e5;
      result.ageHours = Number(ageHours.toFixed(2));
      result.fresh = Number.isFinite(ageHours) && ageHours <= maxAgeHours;
    }

    if (result.present && result.fresh) {
      result.status = 'healthy';
      if (!report.lastSuccessfulSync || Date.parse(result.updatedAt) > Date.parse(report.lastSuccessfulSync)) {
        report.lastSuccessfulSync = result.updatedAt;
      }
    } else if (result.present) {
      result.status = 'stale';
      result.cureSuggestion = `Secret ${secretName} em ${repository} está stale. Execute o workflow N07 Soul Mesh Secret Synchronization.`;
      report.resyncSuggested = true;
    } else {
      result.status = 'missing';
      result.cureSuggestion = `Secret ${secretName} ausente em ${repository}. Execute a sincronização e valide novamente.`;
      report.resyncSuggested = true;
    }
  } catch (error) {
    result.status = 'error';
    result.error = error instanceof Error ? error.message : String(error);
    result.cureSuggestion = 'Verifique permissões do REPO_SECRETS_ADMIN_TOKEN, acesso ao repositório e API de Actions secrets.';
    report.resyncSuggested = true;
  }

  report.repositories.push(result);
  console.log(JSON.stringify(result));
}

if (report.resyncSuggested) {
  report.cureSuggestion = 'Divergência detectada: executar novamente N07 Soul Mesh Secret Synchronization antes de considerar a malha autenticada.';
} else {
  report.cureSuggestion = 'Nenhuma divergência de presença/frescor detectada.';
}

await mkdir('.diagnostics', { recursive: true });
await writeFile('.diagnostics/secret-validator-report.json', JSON.stringify(report, null, 2));

const summaryFile = process.env.GITHUB_STEP_SUMMARY;
if (summaryFile) {
  const unhealthy = report.repositories.filter((item) => item.status !== 'healthy');
  const lines = [
    '# SOUL Mesh Secret Validation',
    '',
    `- Secret: \`${secretName}\``,
    `- Last successful metadata update observed: \`${report.lastSuccessfulSync || 'none'}\``,
    `- Freshness window: **${maxAgeHours}h**`,
    `- Status: **${unhealthy.length === 0 ? 'HEALTHY' : 'ACTION REQUIRED'}**`,
    '',
  ];

  if (unhealthy.length > 0) {
    lines.push('## Cure Suggestion', '', report.cureSuggestion, '');
    for (const item of unhealthy) {
      lines.push(`- **${item.repository}**: ${item.cureSuggestion || item.error || 'unhealthy'}`);
    }
  }

  await writeFile(summaryFile, lines.join('\n') + '\n', { flag: 'a' });
}

const unhealthy = report.repositories.filter((item) => item.status !== 'healthy');
if (unhealthy.length > 0) {
  throw new Error(`Secret validator found ${unhealthy.length} unhealthy repository secret(s); resynchronization is suggested`);
}
