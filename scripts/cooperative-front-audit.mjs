#!/usr/bin/env node
import fs from "node:fs";

class RateLimitError extends Error {}

const eventPath = process.env.GITHUB_EVENT_PATH;
if (!eventPath || !fs.existsSync(eventPath)) {
  console.log("COOPERATIVE_OVERLAP_STATUS=skipped");
  process.exit(0);
}
const event = JSON.parse(fs.readFileSync(eventPath, "utf8"));
const current = event.pull_request;
if (!current) {
  console.log("COOPERATIVE_OVERLAP_STATUS=skipped");
  process.exit(0);
}
const token = process.env.GITHUB_TOKEN;
const repository = process.env.GITHUB_REPOSITORY;
if (!token || !repository) {
  console.log("COOPERATIVE_OVERLAP_STATUS=skipped");
  process.exit(0);
}
const [owner, repo] = repository.split("/");
const api = async (path) => {
  const response = await fetch("https://api.github.com" + path, {
    headers: {
      Accept: "application/vnd.github+json",
      Authorization: "Bearer " + token,
      "X-GitHub-Api-Version": "2022-11-28",
    },
  });
  const body = await response.text();
  if (!response.ok) {
    if (response.status === 403 && /rate limit exceeded/i.test(body)) {
      throw new RateLimitError(`GitHub API rate limit exceeded for ${path}`);
    }
    throw new Error(`GitHub API ${response.status}: ${body}`);
  }
  return JSON.parse(body);
};

try {
  const changed = await api(`/repos/${owner}/${repo}/pulls/${current.number}/files?per_page=100`);
  const currentFiles = new Set(changed.map(file => file.filename));
  const open = await api(`/repos/${owner}/${repo}/pulls?state=open&per_page=100`);
  const overlaps = [];
  for (const pr of open.filter(item => item.number !== current.number)) {
    const files = await api(`/repos/${owner}/${repo}/pulls/${pr.number}/files?per_page=100`);
    const shared = [...new Set(files.map(file => file.filename).filter(file => currentFiles.has(file)))].sort();
    if (shared.length > 0) overlaps.push({ number: pr.number, title: pr.title, files: shared });
  }
  console.log(JSON.stringify({ current_pr: current.number, overlaps }, null, 2));
  console.log(`COOPERATIVE_OVERLAP_STATUS=${overlaps.length ? "attention_required" : "clean"}`);
} catch (error) {
  if (error instanceof RateLimitError) {
    console.log(JSON.stringify({
      current_pr: current.number,
      overlaps: null,
      evidence_state: "BLOCKED_RATE_LIMIT",
      reason: error.message,
    }, null, 2));
    console.log("COOPERATIVE_OVERLAP_STATUS=blocked_rate_limit");
    process.exit(0);
  }
  throw error;
}
