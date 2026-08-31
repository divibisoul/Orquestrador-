package raft

import (
	"testing"
	"time"
)

type machine struct{ applied int }
func (m *machine) Apply([]byte) error { m.applied++; return nil }

func TestElectionAndAppend(t *testing.T) {
	m := &machine{}
	n := NewNode(1, 3, m)
	req := n.StartElection(time.Now())
	if req.Term != 1 { t.Fatalf("term=%d", req.Term) }
	if !n.RecordVote(VoteResponse{Term:1, Granted:true}) { t.Fatal("expected majority with self + peer") }
	entry, err := n.AppendLocal([]byte("cmd"))
	if err != nil { t.Fatal(err) }
	if entry.Index != 1 { t.Fatalf("index=%d", entry.Index) }
}

func TestFollowerAppendAndCommit(t *testing.T) {
	m := &machine{}
	n := NewNode(2, 3, m)
	resp := n.HandleAppend(AppendEntries{Term:1, LeaderID:1, Entries:[]LogEntry{{Term:1, Index:1, Command:[]byte("x")}}, LeaderCommit:1})
	if !resp.Success { t.Fatal("append rejected") }
	if err := n.ApplyCommitted(); err != nil { t.Fatal(err) }
	if m.applied != 1 { t.Fatalf("applied=%d", m.applied) }
}
