package mesh

import (
	"testing"
	"time"
)

func TestMarkStaleMarksExpiredNodes(t *testing.T) {
	r := NewRegistry()
	if err := r.Announce(Node{ID: "n1", Status: "ready", Capabilities: []string{"compute"}, LastHeartbeat: time.Now().Add(-time.Hour)}); err != nil { t.Fatal(err) }
	if got := r.MarkStale(5 * time.Minute); got != 1 { t.Fatalf("expected one stale node, got %d", got) }
	if nodes := r.Snapshot(); len(nodes) != 1 || nodes[0].Status != "stale" { t.Fatalf("expected stale node, got %#v", nodes) }
}

func TestHeartbeatRefreshesNode(t *testing.T) {
	r := NewRegistry()
	if err := r.Announce(Node{ID: "n1", Status: "stale"}); err != nil { t.Fatal(err) }
	before := time.Now()
	if err := r.Heartbeat("n1"); err != nil { t.Fatal(err) }
	nodes := r.Snapshot()
	if len(nodes) != 1 || nodes[0].Status != "ready" || nodes[0].LastHeartbeat.Before(before) { t.Fatalf("heartbeat did not refresh node: %#v", nodes) }
}
