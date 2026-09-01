package mesh

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/divibisoul/Orquestrador-/protocol"
)

func (p *PeerClient) CallWithCorrelation(ctx context.Context, nucleus, capability string, payload map[string]any, correlation string) (map[string]any, error) {
    if ctx == nil { return nil, errors.New("context is nil") }
    nucleus = strings.TrimSpace(nucleus)
    capability = strings.TrimSpace(capability)
    correlation = strings.TrimSpace(correlation)
    if nucleus == "" || capability == "" || correlation == "" { return nil, errors.New("nucleus, capability and correlation are required") }
    if nucleus == protocol.N07 { return nil, errors.New("N07 cannot call itself through peer transport") }
    p.mu.RLock(); peer, ok := p.peers[nucleus]; p.mu.RUnlock()
    if !ok { return nil, fmt.Errorf("peer not configured: %s", nucleus) }
    if p.secret == "" { return nil, errors.New("SOUL_MESH_HMAC_SECRET is not configured") }
    return p.callWithCorrelation(ctx, peer, capability, payload, correlation)
}

func (p *PeerClient) CallBest(ctx context.Context, capability string, payload map[string]any, correlation string) (map[string]any, string, error) {
    if ctx == nil { return nil, "", errors.New("context is nil") }
    capability = strings.TrimSpace(capability)
    if capability == "" { return nil, "", errors.New("capability is required") }
    peers := p.ConfiguredPeers()
    var lastErr error
    best := ""
    bestLatency := time.Duration(1<<63 - 1)
    for _, peer := range peers {
        if peer.Circuit == CircuitOpen && time.Now().Before(peer.RetryAfter) { continue }
        discovery, err := p.Discover(ctx, peer.Nucleus)
        if err != nil { lastErr = err; continue }
        if !capabilityInList(discovery, capability) { continue }
        if peer.Latency > 0 && peer.Latency < bestLatency { best, bestLatency = peer.Nucleus, peer.Latency }
        if best == "" { best = peer.Nucleus }
    }
    if best == "" {
        if lastErr != nil { return nil, "", lastErr }
        return nil, "", fmt.Errorf("capability not discovered on configured peers: %s", capability)
    }
    result, err := p.CallWithCorrelation(ctx, best, capability, payload, correlation)
    return result, best, err
}

func (p *PeerClient) callWithCorrelation(ctx context.Context, peer PeerInfo, capability string, payload map[string]any, correlation string) (map[string]any, error) {
    env := protocol.MeshEnvelope{Version:protocol.SoulMeshVersion, ContractVersion:protocol.SoulMeshContractVersion, MessageID:protocol.NewTraceID(), Source:protocol.N07, Target:peer.Nucleus, Timestamp:time.Now().UnixMilli(), Nonce:protocol.NewTraceID(), CorrelationID:correlation, Type:"CAPABILITY_REQUEST", Payload:map[string]any{"capability":capability,"payload":payload}}
    if err := protocol.SignHMAC(&env, p.secret); err != nil { return nil, err }
    body, err := protocol.EncodeMesh(env); if err != nil { return nil, err }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, peer.URL+"/api/soul-mesh", bytes.NewReader(body)); if err != nil { return nil, err }
    req.Header.Set("Content-Type","application/json")
    req.Header.Set("X-Soul-Contract-Version",protocol.SoulMeshContractVersion)
    req.Header.Set("X-Soul-Correlation-Id",correlation)
    started := time.Now(); resp, err := p.client.Do(req); latency := time.Since(started)
    if err != nil { p.recordFailure(peer.Nucleus,latency,err.Error(),1); return nil,err }
    defer resp.Body.Close()
    if resp.ContentLength > 1<<20 { p.recordFailure(peer.Nucleus,latency,"peer response exceeds configured limit",1); return nil,errors.New("peer response exceeds configured limit") }
    var result map[string]any
    if err := json.NewDecoder(io.LimitReader(resp.Body,1<<20)).Decode(&result); err != nil { p.recordFailure(peer.Nucleus,latency,err.Error(),1); return nil,err }
    if resp.StatusCode < 200 || resp.StatusCode >= 300 { err:=fmt.Errorf("peer request failed: %s",resp.Status);p.recordFailure(peer.Nucleus,latency,err.Error(),1);return result,err }
    if result["contractVersion"] != protocol.SoulMeshContractVersion { err:=errors.New("mesh response contract version mismatch");p.recordFailure(peer.Nucleus,latency,err.Error(),1);return nil,err }
    if result["correlationId"] != correlation { err:=errors.New("peer correlation mismatch");p.recordFailure(peer.Nucleus,latency,err.Error(),1);return nil,err }
    if err:=verifyResponseHMAC(result,p.secret);err!=nil {p.recordFailure(peer.Nucleus,latency,err.Error(),1);return nil,err}
    p.recordSuccess(peer.Nucleus,latency)
    return result,nil
}

func capabilityInList(discovery map[string]any, target string) bool {
    for _, key := range []string{"capabilities","operations","executableCapabilities"} {
        raw, ok := discovery[key]
        if !ok { continue }
        values, ok := raw.([]any); if !ok { continue }
        for _, item := range values {
            value, ok := item.(string); if !ok { continue }
            value = strings.TrimSpace(strings.SplitN(value,"@",2)[0])
            if value == target { return true }
        }
    }
    return false
}
