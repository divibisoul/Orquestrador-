package memory

import (
	"testing"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

func TestMemoryAdapter(t *testing.T) {
	s := New(superagi.NewRuntime())
	if err := s.Working("a", "", "b"); err != nil { t.Fatal(err) }
	if err := s.Semantic("claim", "true"); err != nil { t.Fatal(err) }
	if err := s.Vector("v", []float64{1,2,3}); err != nil { t.Fatal(err) }
}
