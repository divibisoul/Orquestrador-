import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { validateFoundation } from "./validate-sena-foundation.mjs";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const readJson = (relativePath) => JSON.parse(fs.readFileSync(path.join(ROOT, relativePath), "utf8"));

function fixtures() {
  return {
    contract: readJson("integrations/sena/foundation-contract.json"),
    crosswalk: readJson("integrations/sena/public-source-crosswalk.json"),
    registry: readJson("integrations/external-capabilities.json"),
    route: readJson("integrations/sena/vagus-supergpu-prefrontal-route.json")
  };
}

test("SENA foundation is structurally valid while remaining disabled and SPEC_ONLY", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  assert.deepEqual(validateFoundation(contract, crosswalk, registry, route), []);
  assert.equal(contract.component.architectural_status, "SPEC_ONLY");
  assert.equal(contract.deployment.enabled_by_default, false);
  assert.equal(contract.deployment.production_mode, "BLOCKED");
});

test("validator rejects enabling production or introducing another Mesh", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const mutated = structuredClone(contract);
  mutated.deployment.enabled_by_default = true;
  mutated.transport.additional_mesh_allowed = true;
  const errors = validateFoundation(mutated, crosswalk, registry, route);
  assert.ok(errors.some((error) => error.includes("enabled_by_default")));
  assert.ok(errors.some((error) => error.includes("second Mesh")));
});

test("validator requires tracing, correlation, hashes and bounded deadlines", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const mutated = structuredClone(contract);
  mutated.request_contract.required_fields = ["operation", "payload"];
  const errors = validateFoundation(mutated, crosswalk, registry, route);
  assert.ok(errors.some((error) => error.includes("trace_id")));
  assert.ok(errors.some((error) => error.includes("correlation_id")));
  assert.ok(errors.some((error) => error.includes("input_hash")));
  assert.ok(errors.some((error) => error.includes("deadline_unix_ms")));
});

test("validator rejects source duplicates, missing pins and registry drift", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const duplicated = structuredClone(crosswalk);
  duplicated.current_mainline_source_ids.push(structuredClone(duplicated.current_mainline_source_ids[0]));
  assert.ok(validateFoundation(contract, duplicated, registry, route).some((error) => error.includes("must be unique")));

  const unpinned = structuredClone(registry);
  const source = unpinned.repositories.find((entry) => entry.id === "whisper");
  source.revision = "";
  assert.ok(validateFoundation(contract, crosswalk, unpinned, route).some((error) => error.includes("revision-pinned: whisper")));

  const missing = structuredClone(registry);
  missing.repositories = missing.repositories.filter((entry) => entry.id !== "kokoro");
  assert.ok(validateFoundation(contract, crosswalk, missing, route).some((error) => error.includes("does not contain kokoro")));
});

test("validator preserves explicitly open concurrent fronts and hard governance gates", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const noCorrelation = structuredClone(contract);
  noCorrelation.transport.preserve_correlation_id = false;
  const errors = validateFoundation(noCorrelation, crosswalk, registry, route);
  assert.ok(errors.some((error) => error.includes("correlation_id preservation")));

  const badFronts = structuredClone(crosswalk);
  badFronts.active_integration_fronts[0].observed_state = "MERGED";
  assert.ok(validateFoundation(contract, badFronts, registry, route).some((error) => error.includes("marked open")));

  const noEtR = structuredClone(contract);
  noEtR.hard_gates = noEtR.hard_gates.filter((gate) => !gate.includes("SARA ETR"));
  assert.ok(validateFoundation(noEtR, crosswalk, registry, route).some((error) => error.includes("SARA ETR and JEV")));
});

test("validator rejects route bypasses and ownership duplication", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const mutated = structuredClone(route);
  mutated.route[2].to = "SENA general core";
  mutated.route[4].to = "HortaCore";
  mutated.invariants = mutated.invariants.filter((item) => !item.includes("SuperGPU owns compute admission"));
  const errors = validateFoundation(contract, crosswalk, registry, mutated);
  assert.ok(errors.some((error) => error.includes("canonical path")));
  assert.ok(errors.some((error) => error.includes("SuperGPU")));
});

test("validator requires the existing Vagus event protocol and pinned NATS transport only", () => {
  const { contract, crosswalk, registry, route } = fixtures();
  const mutated = structuredClone(route);
  mutated.transport_binding.protocol = "private.sena.bus.v1";
  mutated.transport_binding.no_second_mesh = false;
  mutated.transport_binding.transport_candidate.revision = "";
  mutated.transport_binding.transport_candidate.role = "new independent request Mesh";
  const errors = validateFoundation(contract, crosswalk, registry, mutated);
  assert.ok(errors.some((error) => error.includes("canonical Vagus event protocol")));
  assert.ok(errors.some((error) => error.includes("prohibit a second Mesh")));
  assert.ok(errors.some((error) => error.includes("pinned to a commit")));
  assert.ok(errors.some((error) => error.includes("not a second Mesh")));
});
