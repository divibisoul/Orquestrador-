export const L7_PANEL_FEDERATION = Object.freeze([
  { id: "system-panel", host: "N01", kind: "state+ui" },
  { id: "engineering-panel", host: "N04", kind: "runtime+ui" },
  { id: "audit-panel", host: "N03", kind: "audit-projection+ui" },
  { id: "asasf-panel", host: "SARA/N06", kind: "governed-remediation+ui" },
  { id: "chat-engine", host: "N05", kind: "delegate-bound-chat" },
  { id: "app-orchestrator", host: "N01", kind: "orchestration-adapter" },
  { id: "genesis", host: "N07", kind: "provenance" },
] as const);
