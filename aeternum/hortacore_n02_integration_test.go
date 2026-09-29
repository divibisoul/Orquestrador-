//go:build integration

package aeternum

import (
  "context"
  "os"
  "strings"
  "testing"

  "github.com/divibisoul/Orquestrador-/backend"
  "github.com/divibisoul/Orquestrador-/mesh"
  "github.com/divibisoul/Orquestrador-/neural"
  "github.com/divibisoul/Orquestrador-/orchestrator"
  "github.com/divibisoul/Orquestrador-/prefrontal"
  "github.com/divibisoul/Orquestrador-/supergpu"
)

func TestN07HortaCoreDelegatesBNCv2CSAEDCRSToN02(t *testing.T) {
  if strings.TrimSpace(os.Getenv("SOUL_MESH_N02_URL")) == "" || strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")) == "" {
    t.Skip("N02 live commissioning requires SOUL_MESH_N02_URL and SOUL_MESH_HMAC_SECRET")
  }

  n, err := neural.New(8, 0.05)
  if err != nil {
    t.Fatal(err)
  }
  c, err := prefrontal.New(0.10, 32)
  if err != nil {
    t.Fatal(err)
  }
  g := supergpu.New(nil)
  g.Discover()
  e, err := orchestrator.New(n, c, g)
  if err != nil {
    t.Fatal(err)
  }
  if err := orchestrator.RegisterSuperGPUOperations(e); err != nil {
    t.Fatal(err)
  }
  if err := orchestrator.RegisterAdvancedOperations(e); err != nil {
    t.Fatal(err)
  }

  sara := backend.NewSARAProxy(backend.DefaultConfig())
  h, err := NewHortaCore(e, sara)
  if err != nil {
    t.Fatal(err)
  }
  peers, err := mesh.NewPeerClient(nil)
  if err != nil {
    t.Fatal(err)
  }
  h.SetPeerClient(peers)

  ctx := context.Background()
  cases := []string{"bnc_v2", "csae", "dcrs"}
  for _, moduleID := range cases {
    result, err := h.Execute(ctx, moduleID, []float64{0.1, 0.2, 0.3, 0.4}, map[string]string{
      "input": "real N07 to N02 cognitive continuity commissioning",
    })
    if err != nil {
      t.Fatalf("%s delegation failed: %v", moduleID, err)
    }
    if result["status"] != "ADAPTER_EXECUTED" {
      t.Fatalf("%s returned unexpected status: %#v", moduleID, result["status"])
    }
    if result["owner"] != "N02" {
      t.Fatalf("%s returned unexpected owner: %#v", moduleID, result["owner"])
    }
    if _, ok := result["result"].(map[string]any); !ok {
      t.Fatalf("%s returned no structured owner result: %#v", moduleID, result["result"])
    }
  }
}
