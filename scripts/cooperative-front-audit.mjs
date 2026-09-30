#!/usr/bin/env node
import fs from "node:fs";

const eventPath = process.env.GITHUB_EVENT_PATH;
if (!eventPath || !fs.existsSync(eventPath)) {
  console.log("COOPERATIVE_OVERLAP: event payload unavailable; audit skipped.");
  process.exit(0);
}

const event = JSON.parse(fs.readFileSync(eventPath, "utf8"));
const current = event.pull_request;
if (!current) {
  console.log("COOPERATIVE_OVERLAP: not a pull_request event; nothing to audit.");
  process.exit(0);
}

const token = process.env.GITHUB_TOKEN;
const repo = process.env.GITHUB_REPOSITORY;
if (!token || !repo) {
  console.log("COOPERATIVE_OVERLAP: GitHub context unavailable; audit skipped.");
  process.exit(0);
}

const api = async (path) => {
  const res = await fetch("https://api.github.com" + path, {
    headers: {
      "Accept": "application/vnd.github+json",
      "Authorization": "Bearer " + token,
      "X-GitHub-Api-Version": "2022-11-28",
    },
  });
  if (!res.ok) throw new Error(`GitHub API ${res.status}: ${await res.text()}`);
  return res.json();
};

const ownerRepo = repo.split("/");
const pullList = await api(`/repos/${ownerRepo[0]}/${ownerRepo[1]}/pulls?state=open&per_page=100`);
const openOthers = pullList.filter((pr) => pr.number !== current.number);

const currentFiles = new Set();
for (let page = 1; ; page++) {
  const files = await api(`/repos/${ownerRepo[0]}/${ownerRepo[1]}/pulls/${current.number}/files?per_page=100&page=${page}`);
  if (!files.length) break;
  for (const file of files) currentFiles.add(file.filename);
  if (files.length < 100) break;
}

const overlaps = [];
for (const pr of openOthers) {
  const files = await api(`/repos/${ownerRepo[0]}/${ownerRepo[1]}/pulls/${pr.number}/files?per_page=100`);
  const shared = files.map((f) => f.filename).filter((f) => currentFiles.has(f));
  if (shared.length) {
    overlaps.push({ pr: pr.number, title: pr.title, files: [...new Set(shared)].sort() });
  }
}

console.log(JSON.stringify({
  type: "cooperative-overlap-audit",
  repository: repo,
  current_pr: current.number,
  overlaps,
}, null, 2));

if (overlaps.length === 0) {
  console.log("COOPERATIVE_OVERLAP_STATUS=clean");
} else {
  console.log("COOPERATIVE_OVERLAP_STATUS=attention_required");
  for (const item of overlaps) {
    console.log(`PR #${item.pr} shares: ${item.files.join(", ")}`);
  }
}
