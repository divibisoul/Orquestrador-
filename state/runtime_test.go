package state

import "testing"

func TestStoreDeleteInvalidatesStaleCASVersion(t *testing.T) {
	s := NewStore()
	version := s.Put("k", []byte("value"))
	s.Delete("k")
	_, tombstoneVersion, ok := s.Get("k")
	if ok || tombstoneVersion <= version {
		t.Fatalf("tombstone version=%d ok=%v previous=%d", tombstoneVersion, ok, version)
	}
	if s.CompareAndSwap("k", version, []byte("stale")) {
		t.Fatal("stale CAS unexpectedly succeeded after delete")
	}
	newVersion := s.Put("k", []byte("new"))
	if newVersion <= tombstoneVersion {
		t.Fatalf("new version=%d not greater than tombstone=%d", newVersion, tombstoneVersion)
	}
}
