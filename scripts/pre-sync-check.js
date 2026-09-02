import { mkdir, writeFile } from 'node:fs/promises';

const required = ['SOUL_MESH_TOKEN', 'REPO_SECRETS_ADMIN_TOKEN'];
const status = Object.fromEntries(required.map((name) => [
  name,
  process.env[name]?.trim() ? 'present' : 'absent',
]));

const report = {
  generatedAt: new Date().toISOString(),
  status,
  blocked: Object.values(status).some((value) => value === 'absent'),
};

await mkdir('.diagnostics', { recursive: true });
await writeFile('.diagnostics/pre-sync-status.json', JSON.stringify(report, null, 2) + '\n');

for (const [name, value] of Object.entries(status)) {
  console.log(`${name}=${value}`);
}

if (report.blocked) {
  throw new Error(
    'BLOCKED: Missing required secrets. Configure SOUL_MESH_TOKEN and REPO_SECRETS_ADMIN_TOKEN in N07 repository settings before running sync.',
  );
}
