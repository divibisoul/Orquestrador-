import { execFileSync } from 'node:child_process';
import fs from 'node:fs';

const manifest = JSON.parse(fs.readFileSync(new URL('../soul-nuclei.json', import.meta.url), 'utf8'));
const entries = [
  ...Object.entries(manifest.nuclei).map(([nucleus, spec]) => ({ nucleus, ...spec })),
  manifest.transversal?.SARA ? { nucleus: 'SARA', ...manifest.transversal.SARA } : null,
].filter(Boolean);

const failures = [];

for (const entry of entries) {
  const ref = String(entry.sourceRef ?? '');
  if (ref === 'release-identity:self' || entry.nucleus === 'N07') continue;
  const repository = String(entry.repository ?? '');
  const expected = ref.trim();
  if (!repository || !/^[0-9a-f]{40}$/.test(expected)) {
    failures.push({ nucleus: entry.nucleus, reason: 'INVALID_SOURCE_REF', repository, sourceRef: expected });
    continue;
  }

  let actual = '';
  try {
    actual = execFileSync(
      'git',
      ['ls-remote', 'https://github.com/' + repository + '.git', 'refs/heads/main'],
      { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] },
    ).trim().split(/\s+/)[0];
  } catch (error) {
    failures.push({
      nucleus: entry.nucleus,
      reason: 'REMOTE_MAIN_UNREADABLE',
      repository,
      sourceRef: expected,
      error: error instanceof Error ? error.message : String(error),
    });
    continue;
  }

  if (actual !== expected) failures.push({
    nucleus: entry.nucleus,
    reason: 'SOURCE_REF_DRIFT',
    repository,
    expected,
    actual,
  });
}

const report = {
  status: failures.length === 0 ? 'PASS' : 'FAIL',
  checked: entries.length,
  failures,
  verifiedAt: new Date().toISOString(),
};

console.log(JSON.stringify(report, null, 2));
if (failures.length > 0) process.exit(1);
