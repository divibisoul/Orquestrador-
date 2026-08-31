package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDurableStoreSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	s, err := OpenDurable(path); if err != nil { t.Fatal(err) }
	v, err := s.Put("k", []byte("value")); if err != nil || v != 1 { t.Fatalf("v=%d err=%v",v,err) }
	if err := s.Delete("missing"); err != nil { t.Fatal(err) }
	s2, err := OpenDurable(path); if err != nil { t.Fatal(err) }
	got, ver, ok := s2.Get("k")
	if !ok || string(got)!="value" || ver!=1 { t.Fatalf("got=%q ver=%d ok=%v",got,ver,ok) }
	if _,err:=os.Stat(path);err!=nil{t.Fatal(err)}
}
