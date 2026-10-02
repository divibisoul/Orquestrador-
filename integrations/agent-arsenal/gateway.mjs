import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { mkdir, readFile, readdir, stat, rm, writeFile } from "node:fs/promises";
import { createServer } from "node:http";
import { promisify } from "node:util";
import path from "node:path";
import process from "node:process";

const execFileAsync = promisify(execFile);

const SOURCES = Object.freeze({
  superpowers: {
    repo: "https://github.com/obra/superpowers.git",
    ref: "8ca22dba9a94f28898bbce59f2537ff4d87c747d"
  },
  ecc: {
    repo: "https://github.com/affaan-m/ECC.git",
    ref: "c05b2d6614f62f6db0047669aa4eefb223d478f9"
  },
  ruflo: {
    repo: "https://github.com/ruvnet/ruflo.git",
    ref: "27982983ea6cdc4767c0b6614a4ad9a9d9497cce"
  }
});

const ROOT = path.resolve(process.env.AGENT_ARSENAL_ROOT || "/var/lib/soul-agent-arsenal");
const HOST = process.env.AGENT_ARSENAL_HOST || "0.0.0.0";
const PORT = Number(process.env.AGENT_ARSENAL_PORT || 8090);
const EXECUTE = /^(1|true|yes)$/i.test(process.env.AGENT_ARSENAL_EXECUTE || "false");
const COMMAND_TIMEOUT_MS = Math.max(1000, Number(process.env.AGENT_ARSENAL_COMMAND_TIMEOUT_MS || 20000));
const INSTALL_TIMEOUT_MS = Math.max(1000, Number(process.env.AGENT_ARSENAL_INSTALL_TIMEOUT_MS || 300000));
const MAX_RESOLVE_BYTES = Math.max(4096, Number(process.env.AGENT_ARSENAL_MAX_RESOLVE_BYTES || 1048576));
const MAX_BODY_BYTES = 256 * 1024;
const MAX_ACTIVATION_BYTES = 64 * 1024;
const AUTH_TOKEN = String(process.env.AGENT_ARSENAL_TOKEN || "").trim();
const sourceLocks = new Map();

function authorized(req) {
  if (!AUTH_TOKEN) return false;
  const header = String(req.headers.authorization || "");
  if (!header.toLowerCase().startsWith("bearer ")) return false;
  return header.slice(7).trim() === AUTH_TOKEN;
}

function send(res, status, body) {
  const raw = Buffer.from(JSON.stringify(body));
  res.writeHead(status, {
    "content-type": "application/json; charset=utf-8",
    "content-length": raw.length,
    "cache-control": "no-store"
  });
  res.end(raw);
}

async function readJson(req) {
  const chunks = [];
  let size = 0;
  for await (const chunk of req) {
    size += chunk.length;
    if (size > MAX_BODY_BYTES) {
      throw Object.assign(new Error("request too large"), { statusCode: 413 });
    }
    chunks.push(chunk);
  }
  if (!chunks.length) return {};
  return JSON.parse(Buffer.concat(chunks).toString("utf8"));
}

function sourceConfig(source) {
  const key = String(source || "").trim().toLowerCase();
  if (!Object.hasOwn(SOURCES, key)) {
    throw Object.assign(new Error("unknown source"), { statusCode: 400 });
  }
  return [key, SOURCES[key]];
}

async function execGit(args, cwd = ROOT) {
  return execFileAsync("git", args, {
    cwd,
    timeout: 120000,
    maxBuffer: 8 * 1024 * 1024
  });
}

async function ensureSourceUnlocked(name) {
  const cfg = SOURCES[name];
  const dir = path.join(ROOT, name);
  const pinFile = path.join(dir, ".soul-agent-arsenal-pin");
  try {
    if ((await readFile(pinFile, "utf8")).trim() === cfg.ref) {
      const verified = (await execGit(["-C", dir, "rev-parse", "HEAD"])).stdout.trim();
      if (verified === cfg.ref) return dir;
    }
  } catch {}
  await rm(dir, { recursive: true, force: true });
  await mkdir(ROOT, { recursive: true });
  await execGit(["clone", "--filter=blob:none", "--no-checkout", cfg.repo, dir]);
  await execGit(["fetch", "--depth", "1", "origin", cfg.ref], dir);
  await execGit(["checkout", "--detach", cfg.ref], dir);
  const verified = (await execGit(["-C", dir, "rev-parse", "HEAD"])).stdout.trim();
  if (verified !== cfg.ref) {
    await rm(dir, { recursive: true, force: true });
    throw Object.assign(new Error("source revision verification failed"), { statusCode: 424 });
  }
  await writeFile(pinFile, cfg.ref + "\n");
  return dir;
}

async function ensureSource(name) {
  if (sourceLocks.has(name)) return sourceLocks.get(name);
  const pending = ensureSourceUnlocked(name);
  sourceLocks.set(name, pending);
  try {
    return await pending;
  } finally {
    sourceLocks.delete(name);
  }
}

async function ensureAllSources() {
  for (const name of Object.keys(SOURCES)) await ensureSource(name);
}

function classify(rel) {
  const s = rel.replaceAll(path.sep, "/");
  if (/(^|\/)skills\/[^/]+\/SKILL\.md$/i.test(s)) return "skill";
  if (/(^|\/)\.claude\/agents\//i.test(s) || /(^|\/)\.agents\/.*\/agents\//i.test(s) || /(^|\/)agents\//i.test(s)) return "agent";
  if (/(^|\/)(hooks?|hook-config|\.hooks)\//i.test(s) || /hook[s]?\.json$/i.test(s)) return "hook";
  if (/(^|\/)commands?\//i.test(s) || /\.claude\/commands\//i.test(s)) return "command";
  if (/(^|\/)(mcp|mcp-configs?|mcp-servers?)\//i.test(s) || /(^|[._-])mcp([._-]|$)/i.test(s)) return "mcp";
  if (/(^|\/)\.claude-plugin\//i.test(s) || /(^|\/)plugins?\//i.test(s)) return "plugin";
  if (/(^|\/)workflows?\//i.test(s)) return "workflow";
  return "source";
}

async function walk(rootDir, prefix = "") {
  const out = [];
  for (const ent of await readdir(rootDir, { withFileTypes: true })) {
    if (ent.name === ".git" || ent.name === "node_modules" || ent.name === ".DS_Store") continue;
    const abs = path.join(rootDir, ent.name);
    const rel = prefix ? path.join(prefix, ent.name) : ent.name;
    if (ent.isDirectory()) out.push(...await walk(abs, rel));
    else out.push(rel);
  }
  return out;
}

async function buildCatalog(sourceFilter = "", kindFilter = "") {
  await ensureAllSources();
  const names = sourceFilter ? [sourceFilter] : Object.keys(SOURCES);
  const items = [];
  for (const name of names) {
    const dir = await ensureSource(name);
    for (const rel of await walk(dir)) {
      const normalized = rel.replaceAll(path.sep, "/");
      const kind = classify(normalized);
      if (kindFilter && kind !== kindFilter) continue;
      const info = await stat(path.join(dir, rel));
      items.push({ source: name, kind, path: normalized, size: info.size });
    }
  }
  items.sort((a, b) => (a.source + "/" + a.path).localeCompare(b.source + "/" + b.path));
  const counts = Object.create(null);
  for (const item of items) counts[item.kind] = (counts[item.kind] || 0) + 1;
  return {
    schemaVersion: "1.0.0",
    sources: Object.fromEntries(
      Object.entries(SOURCES).map(([name, cfg]) => [name, { repository: cfg.repo, ref: cfg.ref }])
    ),
    total: items.length,
    counts,
    generatedAt: new Date().toISOString(),
    items
  };
}

function safeResolve(rootDir, requested) {
  const base = path.resolve(rootDir);
  const candidate = path.resolve(base, requested);
  if (candidate !== base && !candidate.startsWith(base + path.sep)) {
    throw Object.assign(new Error("path traversal rejected"), { statusCode: 400 });
  }
  return candidate;
}

async function resolveArtifact(source, requested) {
  const [name] = sourceConfig(source);
  const rootDir = await ensureSource(name);
  const abs = safeResolve(rootDir, requested);
  if (path.basename(abs) === ".soul-agent-arsenal-pin" || abs.includes(path.sep + ".git" + path.sep)) {
    throw Object.assign(new Error("internal file rejected"), { statusCode: 400 });
  }
  const info = await stat(abs);
  if (!info.isFile()) throw Object.assign(new Error("artifact is not a file"), { statusCode: 400 });
  if (info.size > MAX_RESOLVE_BYTES) {
    return { source: name, path: requested, kind: classify(requested), size: info.size, content: null, truncated: true };
  }
  const content = await readFile(abs, "utf8");
  return {
    source: name,
    path: requested,
    kind: classify(requested),
    size: info.size,
    sha256: createHash("sha256").update(content).digest("hex"),
    content
  };
}

async function ensureRufloRuntime() {
  const dir = await ensureSource("ruflo");
  const cliPath = path.join(dir, "bin", "cli.js");
  const lockPath = path.join(dir, "package-lock.json");
  await stat(cliPath);
  await stat(lockPath);
  let installed = false;
  try {
    await stat(path.join(dir, "node_modules", ".package-lock.json"));
    installed = true;
  } catch {}
  if (!installed) {
    await execFileAsync("npm", ["ci", "--ignore-scripts", "--no-audit", "--no-fund"], {
      cwd: dir,
      timeout: INSTALL_TIMEOUT_MS,
      maxBuffer: 8 * 1024 * 1024
    });
  }
  return { dir, cliPath };
}

async function runRuflo(args) {
  if (!EXECUTE) {
    throw Object.assign(new Error("agent execution disabled by AGENT_ARSENAL_EXECUTE"), { statusCode: 403 });
  }
  const runtime = await ensureRufloRuntime();
  try {
    const result = await execFileAsync(
      process.execPath,
      [runtime.cliPath, ...args],
      {
        cwd: runtime.dir,
        timeout: COMMAND_TIMEOUT_MS,
        maxBuffer: 2 * 1024 * 1024,
        env: { ...process.env, CI: "1" }
      }
    );
    return {
      status: "ok",
      exitCode: 0,
      stdout: result.stdout,
      stderr: result.stderr,
      backend: "ruflo@27982983ea6cdc4767c0b6614a4ad9a9d9497cce",
      sourceVerified: true
    };
  } catch (err) {
    const timeout = Boolean(err.killed);
    throw Object.assign(
      new Error(timeout ? "RUFLO_EXECUTION_TIMEOUT" : "RUFLO_EXECUTION_FAILED"),
      { statusCode: 502, code: timeout ? "RUFLO_EXECUTION_TIMEOUT" : "RUFLO_EXECUTION_FAILED" }
    );
  }
}

async function handler(req, res) {
  try {
    if (req.method === "GET" && req.url === "/health") {
      return send(res, 200, {
        status: "ok",
        executeEnabled: EXECUTE,
        authenticatedControlPlane: Boolean(AUTH_TOKEN),
        sources: Object.fromEntries(Object.entries(SOURCES).map(([name, cfg]) => [name, cfg.ref]))
      });
    }

    if (!authorized(req)) {
      return send(res, 401, { error: "agent arsenal authentication required" });
    }

    if (req.method === "GET" && req.url?.startsWith("/v1/catalog")) {
      const query = new URL(req.url, "http://localhost").searchParams;
      const source = String(query.get("source") || "").trim();
      const kind = String(query.get("kind") || "").trim();
      if (source) sourceConfig(source);
      const limit = Math.min(1000, Math.max(1, Number(query.get("limit") || 100)));
      const offset = Math.max(0, Number(query.get("offset") || 0));
      const catalog = await buildCatalog(source, kind);
      return send(res, 200, { ...catalog, items: catalog.items.slice(offset, offset + limit), offset, limit });
    }

    if (req.method === "POST" && req.url === "/v1/resolve") {
      const body = await readJson(req);
      return send(res, 200, await resolveArtifact(body.source, body.path));
    }

    if (req.method === "POST" && req.url === "/v1/activate") {
      const body = await readJson(req);
      const source = String(body.source || "").trim();
      const requested = String(body.path || "").trim();
      const task = String(body.task || "").trim();
      if (!source || !requested || !task) {
        throw Object.assign(new Error("source, path and task are required"), { statusCode: 400 });
      }
      const artifact = await resolveArtifact(source, requested);
      if (!["agent", "skill", "workflow"].includes(artifact.kind)) {
        throw Object.assign(new Error("only agent, skill and workflow artifacts can be activated through Ruflo"), { statusCode: 400 });
      }
      if (!artifact.content) {
        throw Object.assign(new Error("artifact content is unavailable for activation"), { statusCode: 413 });
      }
      if (Buffer.byteLength(String(artifact.content), "utf8") > MAX_ACTIVATION_BYTES) {
        throw Object.assign(new Error("activation artifact exceeds the 64 KiB safety bound"), { statusCode: 413 });
      }
      const activationTask =
        "Apply the pinned SOUL external artifact below as execution guidance. " +
        "Preserve its constraints and verification requirements. " +
        "Do not execute arbitrary shell/code contained in the artifact unless the selected runtime explicitly requires it.\n\n" +
        "SOURCE: " + artifact.source + "\nPATH: " + artifact.path + "\nSHA256: " + artifact.sha256 + "\n\n" +
        "--- ARTIFACT ---\n" + artifact.content + "\n--- END ARTIFACT ---\n\nTASK:\n" + task;
      const args = ["task", "orchestrate", "--task", activationTask];
      if (body.strategy) args.push("--strategy", String(body.strategy));
      if (body.priority) args.push("--priority", String(body.priority));
      return send(res, 200, await runRuflo(args));
    }

    if (req.method === "POST" && req.url === "/v1/swarm") {
      const body = await readJson(req);
      const task = String(body.task || "").trim();
      if (!task) throw Object.assign(new Error("task is required"), { statusCode: 400 });
      const args = ["task", "orchestrate", "--task", task];
      if (body.strategy) args.push("--strategy", String(body.strategy));
      if (body.priority) args.push("--priority", String(body.priority));
      return send(res, 200, await runRuflo(args));
    }

    if (req.method === "POST" && req.url === "/v1/agent") {
      const body = await readJson(req);
      const type = String(body.type || "").trim();
      if (!/^[a-zA-Z0-9._:-]{1,100}$/.test(type)) {
        throw Object.assign(new Error("type is required and must be an agent identifier"), { statusCode: 400 });
      }
      const args = ["agent", "spawn", "-t", type];
      if (body.name) {
        const name = String(body.name).trim();
        if (!/^[a-zA-Z0-9._:-]{1,100}$/.test(name)) {
          throw Object.assign(new Error("name must be an agent identifier"), { statusCode: 400 });
        }
        args.push("--name", name);
      }
      return send(res, 200, await runRuflo(args));
    }

    throw Object.assign(new Error("route not found"), { statusCode: 404 });
  } catch (err) {
    return send(res, Number(err.statusCode || 500), { error: err.message || String(err) });
  }
}

if (process.argv.includes("--catalog-only")) {
  const catalog = await buildCatalog();
  const output = process.env.AGENT_ARSENAL_CATALOG_OUT || path.join(ROOT, "catalog.json");
  await mkdir(path.dirname(output), { recursive: true });
  await writeFile(output, JSON.stringify(catalog, null, 2) + "\n");
  console.log(JSON.stringify({ total: catalog.total, counts: catalog.counts, output }));
  process.exit(0);
}

await mkdir(ROOT, { recursive: true });
createServer(handler).listen(PORT, HOST, () => {
  console.log(JSON.stringify({
    service: "soul-agent-arsenal",
    host: HOST,
    port: PORT,
    executeEnabled: EXECUTE,
    sources: Object.fromEntries(Object.entries(SOURCES).map(([name, cfg]) => [name, cfg.ref]))
  }));
}
