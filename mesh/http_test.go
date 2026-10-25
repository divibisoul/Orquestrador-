package mesh

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/divibisoul/Orquestrador-/neural"
	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/prefrontal"
	"github.com/divibisoul/Orquestrador-/protocol"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef"

type testWire struct {
	Protocol        string         `json:"protocol"`
	ContractVersion string         `json:"contractVersion"`
	ID              string         `json:"id"`
	CorrelationID   string         `json:"correlationId"`
	Source          string         `json:"source"`
	Target          string         `json:"target"`
	Kind            string         `json:"kind"`
	Capability      string         `json:"capability"`
	Payload         map[string]any `json:"payload"`
	Timestamp       int64          `json:"timestamp"`
	Nonce           string         `json:"nonce"`
	HMAC            string         `json:"hmac"`
}

func newTestGateway(t *testing.T) *HTTPGateway {
	t.Helper()
	t.Setenv("N07_MESH_HMAC_SECRET", "")
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "true")
	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpuCompat.New(nil)
	g.Discover()
	e, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	return NewHTTPGateway(e)
}
func canonicalRequest(kind, capability, correlation string, values []float64) map[string]any {
	payload := map[string]any{}
	if values != nil {
		payload["values"] = values
	}
	return map[string]any{"protocol": "soul-mesh/1", "contractVersion": "1.1.0", "id": protocol.NewTraceID(), "correlationId": correlation, "source": "N01", "target": "N07", "kind": kind, "capability": capability, "payload": payload, "timestamp": time.Now().UnixMilli(), "nonce": protocol.NewTraceID()}
}
func TestCanonicalN01BytesOmitsUndefinedTransportButKeepsMeta(t *testing.T) {
	wire := canonicalWireEnvelope{
		Protocol:        "soul-mesh/1",
		ContractVersion: "1.1.0",
		ID:              "message-1",
		CorrelationID:   "correlation-1",
		Source:          "N07",
		Target:          "N01",
		Kind:            "request",
		Capability:      "mesh.describe",
		Payload:         map[string]any{},
		Timestamp:       1790000000000,
		Nonce:           "nonce-1",
	}

	raw, err := canonicalN01Bytes(wire, wire.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	var encoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &encoded); err != nil {
		t.Fatal(err)
	}
	if _, exists := encoded["transport"]; exists {
		t.Fatalf("undefined transport must be omitted like Node JSON.stringify, got %s", raw)
	}
	if string(encoded["meta"]) != "null" {
		t.Fatalf("absent metadata must remain explicit null, got %s", encoded["meta"])
	}

	wire.Meta = map[string]any{"transport": "HTTP"}
	raw, err = canonicalN01Bytes(wire, wire.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	encoded = nil
	if err := json.Unmarshal(raw, &encoded); err != nil {
		t.Fatal(err)
	}
	if string(encoded["transport"]) != "\"HTTP\"" {
		t.Fatalf("defined transport must be preserved, got %s", encoded["transport"])
	}
	if string(encoded["meta"]) != "{\"transport\":\"HTTP\"}" {
		t.Fatalf("metadata must be preserved in canonical bytes, got %s", encoded["meta"])
	}
}

func postWire(t *testing.T, h http.Handler, wire map[string]any) (testWire, int) {
	t.Helper()
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out testWire
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out, rec.Code
}
func TestHTTPGatewayPingAndDescribe(t *testing.T) {
	h := newTestGateway(t)
	ping := canonicalRequest("request", "mesh.ping", "trace-ping", nil)
	got, code := postWire(t, h, ping)
	if code != http.StatusOK || got.CorrelationID != "trace-ping" || got.Source != "N07" || got.Target != "N01" || got.Kind != "response" || got.Capability != "mesh.ping" || got.ContractVersion != "1.1.0" {
		t.Fatalf("unexpected ping response: code=%d envelope=%+v", code, got)
	}
	describe := canonicalRequest("request", "mesh.describe", "trace-describe", nil)
	got, code = postWire(t, h, describe)
	if code != http.StatusOK || got.CorrelationID != "trace-describe" || got.Kind != "response" {
		t.Fatalf("unexpected describe response: code=%d envelope=%+v", code, got)
	}
	operations, ok := got.Payload["executableCapabilities"].([]any)
	if !ok {
		t.Fatalf("mesh.describe must expose live executableCapabilities, got %#v", got.Payload["executableCapabilities"])
	}
	for _, required := range []string{"mesh.describe", "mesh.ping", "neural.forward@1.0.0"} {
		found := false
		for _, raw := range operations {
			if raw == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mesh.describe omitted registered executable capability %q: %#v", required, operations)
		}
	}
	if got.Payload["declaredCapabilities"] == nil || got.Payload["transports"] == nil {
		t.Fatalf("mesh.describe response is missing common discovery fields: %#v", got.Payload)
	}
}
func TestExternalFederationCatalogAndAllThirtyTwoSourcesAreDiscoverableOverMeshHTTP(t *testing.T) {
	h := newTestGateway(t)
	registry, err := orchestrator.NewExternalAdapterRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestrator.RegisterExternalAdapterOperations(h.Engine, registry); err != nil {
		t.Fatal(err)
	}

	// One common bootstrap advertises the actual federation handlers.
	describeWire := canonicalRequest("request", "mesh.describe", "trace-external-mesh-bootstrap", nil)
	got, code := postWire(t, h, describeWire)
	if code != http.StatusOK {
		t.Fatalf("mesh.describe failed over HTTP: code=%d payload=%#v", code, got.Payload)
	}
	executable, ok := got.Payload["executableCapabilities"].([]any)
	if !ok {
		t.Fatalf("N07 mesh.describe omitted executableCapabilities: %#v", got.Payload)
	}
	for _, required := range []string{
		orchestrator.ExternalFederationDescribeOperation,
		orchestrator.ExternalFederationExecuteOperation,
		orchestrator.ExternalFederationSourceDescribeOperation,
		orchestrator.ExternalFederationSourceProbeOperation,
	} {
		found := false
		for _, advertised := range executable {
			if advertised == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("N07 did not advertise registered federation operation %q: %#v", required, executable)
		}
	}

	// Retrieve source identities through the same Mesh route used by peer nuclei.
	catalogWire := canonicalRequest("request", orchestrator.ExternalFederationDescribeOperation, "trace-external-catalog", nil)
	catalogResponse, code := postWire(t, h, catalogWire)
	if code != http.StatusOK || catalogResponse.Kind != "response" {
		t.Fatalf("external federation catalog failed over Mesh HTTP: code=%d envelope=%+v", code, catalogResponse)
	}
	catalogMetadata, ok := catalogResponse.Payload["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("catalog response omitted metadata: %#v", catalogResponse.Payload)
	}
	catalogRaw, ok := catalogMetadata["federation_json"].(string)
	if !ok || strings.TrimSpace(catalogRaw) == "" {
		t.Fatalf("catalog response omitted federation_json: %#v", catalogMetadata)
	}
	var catalog struct {
		State string `json:"state"`
		RuntimeEvidence string `json:"runtime_evidence"`
		ProviderCount int `json:"provider_count"`
		ExpansionProviderCount int `json:"expansion_provider_count"`
		DeclaredExternalRepositoryCount int `json:"declared_external_repository_count"`
		Providers []struct { ID string `json:"id"` } `json:"providers"`
		ExpansionProviders []struct {
			ID string `json:"id"`
			State string `json:"state"`
			ExecutionEnabled bool `json:"executionEnabled"`
			ExecutionProven bool `json:"executionProven"`
			RuntimeRouteOperations []string `json:"runtimeRouteOperations"`
		} `json:"expansion_providers"`
	}
	if err := json.Unmarshal([]byte(catalogRaw), &catalog); err != nil {
		t.Fatalf("catalog response contains invalid JSON: %v", err)
	}
	if catalog.State != "PASS" || catalog.RuntimeEvidence != "UNVERIFIED" ||
		catalog.ProviderCount != 22 || catalog.ExpansionProviderCount != 10 ||
		catalog.DeclaredExternalRepositoryCount != 32 ||
		len(catalog.Providers) != 22 || len(catalog.ExpansionProviders) != 10 {
		t.Fatalf("catalog inventory/evidence mismatch: %#v", catalog)
	}

	ids := make([]string, 0, 32)
	seen := make(map[string]struct{}, 32)
	for _, provider := range catalog.Providers {
		ids = append(ids, provider.ID)
	}
	for _, provider := range catalog.ExpansionProviders {
		ids = append(ids, provider.ID)
	}
	projectedIDs := make(map[string]struct{}, len(catalog.ExpansionProviders))
	for _, projected := range catalog.ExpansionProviders {
		projectedIDs[projected.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate public upstream in Mesh catalog: %s", id)
		}
		seen[id] = struct{}{}
		wire := canonicalRequest("request", orchestrator.ExternalFederationSourceDescribeOperation, "trace-source-"+id, nil)
		wire["payload"].(map[string]any)["provider"] = id
		response, status := postWire(t, h, wire)
		if status != http.StatusOK || response.Kind != "response" {
			t.Fatalf("source %s was not discoverable over Mesh HTTP: status=%d envelope=%+v", id, status, response)
		}
		sourceMetadata, ok := response.Payload["metadata"].(map[string]any)
		if !ok {
			t.Fatalf("source %s omitted HTTP Mesh result metadata: %#v", id, response.Payload)
		}
		sourceRaw, ok := sourceMetadata["external_result_json"].(string)
		if !ok || strings.TrimSpace(sourceRaw) == "" {
			t.Fatalf("source %s omitted canonical source evidence: %#v", id, sourceMetadata)
		}
		var evidence struct {
			State string `json:"state"`
			Provider string `json:"provider"`
			Owner string `json:"owner"`
			ExpectedRevision string `json:"expectedRevision"`
			ExecutionEnabled bool `json:"executionEnabled"`
			ExecutionProven bool `json:"executionProven"`
			RuntimeEvidence string `json:"runtimeEvidence"`
			MeshMetadataOperations []string `json:"meshMetadataOperations"`
			RuntimeRouteOperations []string `json:"runtimeRouteOperations"`
		}
		if err := json.Unmarshal([]byte(sourceRaw), &evidence); err != nil {
			t.Fatalf("source %s returned malformed evidence: %v; raw=%s", id, err, sourceRaw)
		}
		if evidence.Provider != id || evidence.Owner == "" || evidence.ExpectedRevision == "" ||
			evidence.RuntimeEvidence != "UNVERIFIED" || evidence.ExecutionProven ||
			(evidence.State != "PROJECTED" && evidence.State != "UNMEASURABLE" && evidence.State != "BLOCKED") ||
			len(evidence.MeshMetadataOperations) < 2 {
			t.Errorf("source %s returned incomplete/unsafe evidence over Mesh HTTP: %#v", id, evidence)
		}

	// Keep the 32 descriptions and 32 probes in separate request windows.
	// The gateway's per-source limit remains enforced; this test must not bypass production throttling.
	for _, id := range ids {
		probeWire := canonicalRequest("request", orchestrator.ExternalFederationSourceProbeOperation, "trace-probe-"+id, nil)
		probeWire["payload"].(map[string]any)["provider"] = id
		probeResponse, probeStatus := postWire(t, h, probeWire)
		if probeStatus != http.StatusOK || probeResponse.Kind != "response" {
			t.Fatalf("source %s was not probeable through common Mesh HTTP API: status=%d envelope=%+v", id, probeStatus, probeResponse)
		}
		probeMetadata, ok := probeResponse.Payload["metadata"].(map[string]any)
		if !ok {
			t.Fatalf("source %s probe omitted HTTP result metadata: %#v", id, probeResponse.Payload)
		}
		probeRaw, ok := probeMetadata["external_result_json"].(string)
		if !ok || strings.TrimSpace(probeRaw) == "" {
			t.Fatalf("source %s probe omitted evidence JSON: %#v", id, probeMetadata)
		}
		var probeEvidence struct {
			Provider string `json:"provider"`
			ExecutionProven bool `json:"executionProven"`
			RuntimeEvidence string `json:"runtimeEvidence"`
			SourceProbeState string `json:"sourceProbeState"`
			SourceProbe map[string]any `json:"sourceProbe"`
		}
		if err := json.Unmarshal([]byte(probeRaw), &probeEvidence); err != nil {
			t.Fatalf("source %s returned malformed probe evidence: %v; raw=%s", id, err, probeRaw)
		}
		if probeEvidence.Provider != id || probeEvidence.ExecutionProven ||
			probeEvidence.RuntimeEvidence != "UNVERIFIED" || probeEvidence.SourceProbe == nil ||
			probeEvidence.SourceProbe["source_integrity_only"] != true ||
			probeEvidence.SourceProbe["execution_proven"] != false {
			t.Errorf("source %s returned unsafe/incomplete probe evidence: %#v", id, probeEvidence)
		}
		if _, projected := projectedIDs[id]; projected {
			if probeEvidence.SourceProbe["execution_enabled"] != false ||
				(probeEvidence.SourceProbeState != "PROJECTED" && probeEvidence.SourceProbeState != "BLOCKED") {
				t.Errorf("projected source %s must remain metadata-only: %#v", id, probeEvidence)
			}
		} else if _, ok := probeEvidence.SourceProbe["package_available"].(bool); !ok {
			t.Errorf("active source %s probe must report package availability: %#v", id, probeEvidence)
		}
	}

	if len(seen) != 32 {
		t.Fatalf("Mesh exposed %d unique public upstreams, want 32", len(seen))
	}
	for _, projected := range catalog.ExpansionProviders {
		if projected.State != "PROJECTED" || projected.ExecutionEnabled || projected.ExecutionProven || len(projected.RuntimeRouteOperations) != 0 {
			t.Errorf("projected source %s acquired an execution route: %#v", projected.ID, projected)
		}
	}
}

func TestHTTPGatewayExecutesRealNeuralCapability(t *testing.T) {
	h := newTestGateway(t)
	wire := canonicalRequest("request", "neural.forward", "trace-neural", []float64{1, 2, 3, 4, 5, 6, 7, 8})
	got, code := postWire(t, h, wire)
	if code != http.StatusOK || got.CorrelationID != "trace-neural" || got.Kind != "response" {
		t.Fatalf("unexpected neural response: code=%d envelope=%+v", code, got)
	}
	values, ok := got.Payload["values"].([]any)
	if !ok || len(values) != 8 {
		t.Fatalf("expected eight neural outputs, got %#v", got.Payload["values"])
	}
	if got.Payload["status"] != "ok" {
		t.Fatalf("expected successful execution, got %#v", got.Payload["status"])
	}
}
func TestHTTPGatewayPreservesTraceIdentity(t *testing.T) {
	h := newTestGateway(t)
	wire := canonicalRequest("request", "neural.forward", "trace-preserved", []float64{1, 2, 3, 4, 5, 6, 7, 8})
	got, code := postWire(t, h, wire)
	if code != http.StatusOK || got.CorrelationID != "trace-preserved" {
		t.Fatalf("correlation identity was not preserved: code=%d envelope=%+v", code, got)
	}
}
func TestHTTPGatewayHMACAuthentication(t *testing.T) {
	t.Setenv("N07_MESH_HMAC_SECRET", testSecret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	h := newTestGateway(t)
	h.Secret = testSecret
	wire := canonicalRequest("request", "mesh.ping", "trace-header-hmac", nil)
	raw := canonicalWireEnvelope{Protocol: wire["protocol"].(string), ContractVersion: wire["contractVersion"].(string), ID: wire["id"].(string), CorrelationID: wire["correlationId"].(string), Source: wire["source"].(string), Target: wire["target"].(string), Kind: wire["kind"].(string), Capability: wire["capability"].(string), Payload: wire["payload"].(map[string]any), Timestamp: wire["timestamp"].(int64), Nonce: wire["nonce"].(string)}
	canonical, err := canonicalN01Bytes(raw, raw.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write(canonical)
	wire["hmac"] = hex.EncodeToString(mac.Sum(nil))
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	req.Header.Set("x-soul-mesh-nonce", raw.Nonce)
	req.Header.Set("x-soul-mesh-hmac", wire["hmac"].(string))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected N01-style HMAC request to succeed, got %d: %s", rec.Code, rec.Body.String())
	}
}
func TestHTTPGatewayRejectsHeaderReplay(t *testing.T) {
	t.Setenv("N07_MESH_HMAC_SECRET", testSecret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	h := newTestGateway(t)
	h.Secret = testSecret
	wire := canonicalRequest("request", "mesh.ping", "trace-replay", nil)
	raw := canonicalWireEnvelope{Protocol: wire["protocol"].(string), ContractVersion: wire["contractVersion"].(string), ID: wire["id"].(string), CorrelationID: wire["correlationId"].(string), Source: wire["source"].(string), Target: wire["target"].(string), Kind: wire["kind"].(string), Capability: wire["capability"].(string), Payload: wire["payload"].(map[string]any), Timestamp: wire["timestamp"].(int64), Nonce: wire["nonce"].(string)}
	canonical, err := canonicalN01Bytes(raw, raw.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write(canonical)
	sig := hex.EncodeToString(mac.Sum(nil))
	body, _ := json.Marshal(wire)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
		req.Header.Set("x-soul-mesh-nonce", raw.Nonce)
		req.Header.Set("x-soul-mesh-hmac", sig)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusOK {
			t.Fatalf("first authenticated request failed: %d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusUnauthorized {
			t.Fatalf("replayed request was accepted: %d", rec.Code)
		}
	}
}
func TestHTTPGatewayRejectsWrongContract(t *testing.T) {
	h := newTestGateway(t)
	wire := canonicalRequest("request", "mesh.ping", "trace-bad", nil)
	wire["contractVersion"] = "9.9.9"
	got, code := postWire(t, h, wire)
	if code != http.StatusBadRequest || got.Payload["error"] == nil {
		t.Fatalf("expected contract rejection, code=%d envelope=%+v", code, got)
	}
}
func TestHTTPGatewayRejectsUnauthenticatedWhenLocalModeDisabled(t *testing.T) {
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")
	t.Setenv("N07_MESH_HMAC_SECRET", "")
	h := newTestGateway(t)
	h.AllowUnauthenticatedLocal = false
	wire := canonicalRequest("request", "mesh.ping", "trace-auth", nil)
	_, code := postWire(t, h, wire)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", code)
	}
}
func TestHTTPGatewayHonorsContextCancellation(t *testing.T) {
	h := newTestGateway(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	wire := canonicalRequest("request", "neural.forward", "trace-cancel", []float64{1, 2, 3, 4, 5, 6, 7, 8})
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body)).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("cancelled request must not be reported as a successful request")
	}
}

func TestCopyStructuredExternalFederationMetadata(t *testing.T) {
	metadata := map[string]string{}
	if err := copyStructuredCapabilityMetadata(metadata, "external.federation.source.describe@1.0.0", map[string]any{"provider": " RAVANA "}); err != nil {
		t.Fatal(err)
	}
	if metadata["provider"] != "ravana" {
		t.Fatalf("provider payload was not normalized: %#v", metadata)
	}

	nested := map[string]string{}
	if err := copyStructuredCapabilityMetadata(nested, "external.federation.source.probe@1.0.0", map[string]any{
		"metadata": map[string]any{"provider": "LangGraph"},
	}); err != nil {
		t.Fatal(err)
	}
	if nested["provider"] != "langgraph" {
		t.Fatalf("nested provider metadata was not normalized: %#v", nested)
	}

	execute := map[string]string{}
	if err := copyStructuredCapabilityMetadata(execute, "external.federation.execute@1.0.0", map[string]any{
		"provider": "langgraph", "external_operation": "workflow.invoke",
	}); err != nil {
		t.Fatal(err)
	}
	if execute["provider"] != "langgraph" || execute["external_operation"] != "workflow.invoke" {
		t.Fatalf("external execution metadata was not normalized: %#v", execute)
	}
	if err := copyStructuredCapabilityMetadata(map[string]string{}, "external.federation.source.describe@1.0.0", map[string]any{}); err == nil {
		t.Fatal("missing provider must fail closed")
	}
}

func TestCanonicalExternalFederationSourceCallWorksOverMesh(t *testing.T) {
	n, err := neural.New(4, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 8)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpuCompat.New(nil)
	g.Discover()
	engine, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{
		"external.federation.source.describe@1.0.0",
		"external.federation.source.probe@1.0.0",
	} {
		op := operation
		if err := engine.Register(op, func(ctx context.Context, m protocol.Message) (protocol.Result, error) {
			provider := m.Metadata["provider"]
			if provider != "ravana" {
				return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07", Target: m.Source, Status: "error", Error: "PROVIDER_NOT_NORMALIZED"}, nil
			}
			return protocol.Result{TraceID: m.TraceID, CorrelationID: m.CorrelationID, Source: "N07", Target: m.Source, Status: "ok", Metadata: map[string]string{"provider": provider}}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	// This test targets payload compatibility and cross-capability dispatch; auth
	// policy is covered independently by the gateway HMAC tests above.
	handler := &HTTPGateway{Engine: engine, AllowUnauthenticatedLocal: true}
	for _, capability := range []string{
		"external.federation.source.describe@1.0.0",
		"external.federation.source.probe@1.0.0",
	} {
		wire := canonicalRequest("request", capability, "trace-"+capability, nil)
		wire["payload"].(map[string]any)["provider"] = "RAVANA" // same payload shape used by PeerClient
		got, code := postWire(t, handler, wire)
		if code != http.StatusOK {
			t.Fatalf("Mesh source route %s failed: code=%d payload=%#v", capability, code, got.Payload)
		}
		metadata, ok := got.Payload["metadata"].(map[string]any)
		if !ok || metadata["provider"] != "ravana" {
			t.Fatalf("Mesh source route %s did not propagate normalized provider metadata: %#v", capability, got.Payload)
		}
	}
}

func TestCopyStructuredSuperGPUMetadata(t *testing.T) {
	metadata := map[string]string{}
	err := copyStructuredCapabilityMetadata(metadata, "mesh.supergpu.execute@1.0.0", map[string]any{
		"values": []any{2.0, 3.0},
		"metadata": map[string]any{
			"operation":           "square",
			"device":              "gpu-0",
			"require_accelerator": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if metadata["operation"] != "square" || metadata["device"] != "gpu-0" || metadata["require_accelerator"] != "true" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
}

func TestCopyStructuredSuperAGIFabricMetadata(t *testing.T) {
	metadata := map[string]string{}
	err := copyStructuredCapabilityMetadata(metadata, "superagi.fabric.execute@1.0.0", map[string]any{
		"agent_provider": "crewai",
		"goal":           "prepare accelerated work",
		"roles":          []any{"planner", "executor"},
		"compute_values": []any{2.0, 3.0},
		"metadata": map[string]any{
			"agent_mode":          "multiagent-facade",
			"compute_operation":   "square",
			"require_accelerator": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if metadata["agent_provider"] != "crewai" || metadata["goal"] != "prepare accelerated work" {
		t.Fatalf("unexpected fabric metadata: %#v", metadata)
	}
	if metadata["compute_values_json"] != "[2,3]" {
		t.Fatalf("unexpected compute values: %s", metadata["compute_values_json"])
	}
}


func TestEnhancedFederatedGatewayAuthenticatedPingDoesNotReuseNonce(t *testing.T) {
	t.Setenv("N07_MESH_HMAC_SECRET", testSecret)
	t.Setenv("N07_MESH_ALLOW_UNAUTH_LOCAL", "false")

	n, err := neural.New(8, .05)
	if err != nil {
		t.Fatal(err)
	}
	c, err := prefrontal.New(.10, 32)
	if err != nil {
		t.Fatal(err)
	}
	g := supergpuCompat.New(nil)
	g.Discover()
	engine, err := orchestrator.New(n, c, g)
	if err != nil {
		t.Fatal(err)
	}
	handler := NewEnhancedFederatedHTTPGateway(engine)

	wire := canonicalRequest("request", "mesh.ping", "trace-enhanced-header-hmac", nil)
	raw := canonicalWireEnvelope{
		Protocol: wire["protocol"].(string),
		ContractVersion: wire["contractVersion"].(string),
		ID: wire["id"].(string),
		CorrelationID: wire["correlationId"].(string),
		Source: wire["source"].(string),
		Target: wire["target"].(string),
		Kind: wire["kind"].(string),
		Capability: wire["capability"].(string),
		Payload: wire["payload"].(map[string]any),
		Timestamp: wire["timestamp"].(int64),
		Nonce: wire["nonce"].(string),
	}
	canonical, err := canonicalN01Bytes(raw, raw.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write(canonical)
	signature := hex.EncodeToString(mac.Sum(nil))
	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/soul-mesh", bytes.NewReader(body))
	req.Header.Set("x-soul-mesh-nonce", raw.Nonce)
	req.Header.Set("x-soul-mesh-hmac", signature)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated federated mesh.ping should pass one authentication boundary, got %d: %s", rec.Code, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["correlationId"] != "trace-enhanced-header-hmac" || response["source"] != "N07" || response["target"] != "N01" || response["kind"] != "response" {
		t.Fatalf("unexpected authenticated ping response: %#v", response)
	}
	if err := verifyResponseHMAC(response, testSecret); err != nil {
		t.Fatalf("authenticated ping response HMAC did not validate: %v", err)
	}
}
