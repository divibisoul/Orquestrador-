package memory

import (
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

// Store is an adapter over SuperAGI's canonical Memory store.
type Store struct { Runtime *superagi.Runtime }

func New(r *superagi.Runtime) *Store {
	if r == nil { r = superagi.NewRuntime() }
	return &Store{Runtime:r}
}

func (s *Store) Working(items ...string) error {
	if s == nil || s.Runtime == nil { return errors.New("memory runtime unavailable") }
	clean := make([]string,0,len(items))
	for _, item := range items { if v:=strings.TrimSpace(item); v!="" { clean=append(clean,v) } }
	if len(clean)>0 { s.Runtime.WorkingMemory(clean...) }
	return nil
}
func (s *Store) Episodic(event string) error {
	if s == nil || s.Runtime == nil { return errors.New("memory runtime unavailable") }
	if strings.TrimSpace(event)=="" { return errors.New("event required") }
	s.Runtime.EpisodicMemory(strings.TrimSpace(event)); return nil
}
func (s *Store) Semantic(key, value string) error {
	if s == nil || s.Runtime == nil { return errors.New("memory runtime unavailable") }
	if strings.TrimSpace(key)=="" { return errors.New("key required") }
	s.Runtime.SemanticMemory(strings.TrimSpace(key), value); return nil
}
func (s *Store) Procedural(key, value string) error {
	if s == nil || s.Runtime == nil { return errors.New("memory runtime unavailable") }
	if strings.TrimSpace(key)=="" { return errors.New("key required") }
	s.Runtime.ProceduralMemory(strings.TrimSpace(key), value); return nil
}
func (s *Store) Vector(key string, value []float64) error {
	if s == nil || s.Runtime == nil { return errors.New("memory runtime unavailable") }
	if strings.TrimSpace(key)=="" { return errors.New("key required") }
	s.Runtime.VectorMemory(strings.TrimSpace(key), value); return nil
}
