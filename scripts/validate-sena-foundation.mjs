import fs from "node:fs";
import path from "node:path";

const ROOT = process.cwd();

function readJson(relativePath) {
  return JSON.parse(fs.readFileSync(path.join(ROOT, relativePath), "utf8"));
}

export function validateFoundation(contract, crosswalk, canonicalRegistry) {
  const errors = [];
  const requireValue = (condition, message) => {
    if (!condition) errors.push(message);
  };

  requireValue(contract?.schema_version === "1.0.0", "contract.schema_version must be 1.0.0");
  requireValue(contract?.component?.id === "SENA", "component.id must be SENA");
  requireValue(contract?.component?.architectural_status === "SPEC_ONLY", "SENA must remain SPEC_ONLY until verified runtime evidence exists");
  requireValue(contract?.component?.new_nucleus === false, "SENA must not create a new SOUL nucleus");
  requireValue(contract?.component?.new_mesh === false, "SENA must not create a second Mesh");

  requireValue(contract?.deployment?.enabled_by_default === false, "deployment.enabled_by_default must be false");
  requireValue(contract?.deployment?.initial_mode === "shadow", "initial mode must be shadow");
  requireValue(contract?.deployment?.production_mode === "BLOCKED", "production must remain BLOCKED");
  requireValue(contract?.deployment?.allow_host_output_mutation === false, "shadow/initial mode must not mutate host output");
  requireValue(contract?.deployment?.allow_autonomous_external_actions === false, "autonomous external actions must be disabled");
  requireValue(contract?.deployment?.allow_online_training === false, "online training must be disabled");
  requireValue(contract?.deployment?.health_is_execution_proof === false, "health must not be treated as execution proof");

  requireValue(contract?.transport?.canonical_mesh_contract === "SOUL Mesh 1.1.0", "transport must use canonical SOUL Mesh 1.1.0");
  requireValue(contract?.transport?.authority === "N07", "N07 must remain the transport/federation authority");
  requireValue(contract?.transport?.additional_mesh_allowed === false, "a second Mesh is prohibited");
  requireValue(contract?.transport?.preserve_trace_id === true, "trace_id preservation is required");
  requireValue(contract?.transport?.preserve_correlation_id === true, "correlation_id preservation is required");
  requireValue(contract?.transport?.max_request_bytes === 1048576, "payload limit must remain explicitly bounded to 1 MiB");

  const requestFields = new Set(contract?.request_contract?.required_fields ?? []);
  for (const field of ["trace_id", "correlation_id", "input_hash", "deadline_unix_ms", "mode", "payload"]) {
    requireValue(requestFields.has(field), "request_contract.required_fields must include " + field);
  }

  const responseFields = new Set(contract?.response_contract?.required_fields ?? []);
  for (const field of ["status", "trace_id", "correlation_id", "input_hash", "output_hash", "evidence_state", "source_revision"]) {
    requireValue(responseFields.has(field), "response_contract.required_fields must include " + field);
  }

  const expectedSourceIds = [
    "superpowers", "superagi", "langgraph", "crewai", "microsoft-agent-framework",
    "openhands", "metagpt", "agentscope", "letta-code", "browser-use",
    "smolagents", "pydantic-ai", "llama-index", "dspy", "whisper", "kokoro"
  ];
  const mappedSources = crosswalk?.current_mainline_source_ids ?? [];
  const mappedIds = mappedSources.map((source) => source?.id);
  const uniqueMappedIds = new Set(mappedIds);
  requireValue(mappedIds.length === expectedSourceIds.length, "crosswalk must account for all 16 mainline source IDs in this audit snapshot");
  requireValue(uniqueMappedIds.size === mappedIds.length, "crosswalk source IDs must be unique");
  for (const id of expectedSourceIds) {
    requireValue(uniqueMappedIds.has(id), "crosswalk is missing mainline source " + id);
  }

  const registryEntries = canonicalRegistry?.repositories ?? [];
  const registryById = new Map(registryEntries.map((source) => [source?.id, source]));
  for (const source of mappedSources) {
    const registered = registryById.get(source?.id);
    requireValue(Boolean(registered), "canonical external registry does not contain " + source?.id);
    if (registered) {
      requireValue(typeof registered.revision === "string" && registered.revision.trim().length > 0,
        "canonical registry source must remain revision-pinned: " + source.id);
    }
  }

  const observedPullRequests = [
    ...(crosswalk?.candidate_source_sets_in_open_prs ?? []),
    ...(crosswalk?.active_integration_fronts ?? [])
  ];
  requireValue(observedPullRequests.length > 0, "crosswalk must preserve the active-front snapshot");
  for (const item of observedPullRequests) {
    requireValue(typeof item.pr === "string" && item.pr.includes("#"), "each front must reference a repository and PR number");
    requireValue(typeof item.observed_state === "string" && item.observed_state.startsWith("OPEN"),
      "snapshot fronts must be clearly marked open at inspection time");
  }

  const boundary = contract?.authority_boundaries ?? {};
  for (const owner of ["orchestration_and_federation", "regeneration_governance_provenance", "typed_decision_and_guardrails", "persistence", "compute_admission", "signal_transport", "state_persistence"]) {
    requireValue(typeof boundary[owner] === "string" && boundary[owner].trim().length > 0,
      "authority boundary must explicitly preserve " + owner);
  }

  const gates = new Set(contract?.hard_gates ?? []);
  requireValue([...gates].some((gate) => gate.includes("SARA ETR") && gate.includes("JEV")),
    "decision-affecting output must be gated by SARA ETR and JEV guardrails");
  requireValue([...gates].some((gate) => gate.includes("authenticated end-to-end")),
    "REAL state requires authenticated end-to-end evidence");

  requireValue(contract?.memory_policy?.initial_access === "read_only", "initial memory access must be read-only");
  requireValue(contract?.memory_policy?.no_second_persistent_memory_authority === true, "a second persistent-memory authority is prohibited");
  requireValue(contract?.learning_policy?.allow_online_training === false, "learning policy must explicitly prohibit online training");
  requireValue(contract?.learning_policy?.human_approval_required_for_policy_promotion === true, "policy promotion requires human approval");
  requireValue(contract?.state_semantics?.no_fake_pass === true, "state semantics must prohibit synthetic PASS");

  return errors;
}

function main() {
  try {
    const contract = readJson("integrations/sena/foundation-contract.json");
    const crosswalk = readJson("integrations/sena/public-source-crosswalk.json");
    const registry = readJson("integrations/external-capabilities.json");
    const errors = validateFoundation(contract, crosswalk, registry);
    if (errors.length > 0) {
      for (const error of errors) console.error("FAIL:", error);
      process.exitCode = 1;
      return;
    }
    console.log("PASS: SENA foundation contract, source crosswalk and authority boundaries are internally consistent.");
    console.log("STATUS: SPEC_ONLY; runtime integration, source execution and production remain unverified/blocked.");
  } catch (error) {
    console.error("FAIL: unable to load/validate SENA foundation files:", error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
  }
}

if (import.meta.url === new URL(process.argv[1], "file://").href) {
  main();
}
