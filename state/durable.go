package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type DurableStore struct {
	mu       sync.RWMutex
	path     string
	values   map[string][]byte
	versions map[string]uint64
}

type durableFile struct {
	Values   map[string][]byte `json:"values"`
	Versions map[string]uint64 `json:"versions"`
}

func OpenDurable(path string) (*DurableStore, error) {
	path = filepath.Clean(path)
	if path == "." || path == "" {
		return nil, errors.New("state path required")
	}
	s := &DurableStore{path: path, values: map[string][]byte{}, versions: map[string]uint64{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var file durableFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if file.Values != nil {
		s.values = cloneValues(file.Values)
	}
	if file.Versions != nil {
		s.versions = cloneVersions(file.Versions)
	}
	return s, nil
}

func (s *DurableStore) Put(key string, value []byte) (uint64, error) {
	if s == nil {
		return 0, errors.New("nil durable store")
	}
	if key == "" {
		return 0, errors.New("key required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = append([]byte(nil), value...)
	s.versions[key]++
	if err := s.flushLocked(); err != nil {
		return 0, err
	}
	return s.versions[key], nil
}

func (s *DurableStore) CompareAndSwap(key string, version uint64, value []byte) (bool, error) {
	if s == nil {
		return false, errors.New("nil durable store")
	}
	if key == "" {
		return false, errors.New("key required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.versions[key] != version {
		return false, nil
	}
	s.values[key] = append([]byte(nil), value...)
	s.versions[key]++
	if err := s.flushLocked(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *DurableStore) Get(key string) ([]byte, uint64, bool) {
	if s == nil {
		return nil, 0, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.values[key]
	return append([]byte(nil), v...), s.versions[key], ok
}

func (s *DurableStore) Delete(key string) error {
	if s == nil {
		return errors.New("nil durable store")
	}
	if key == "" {
		return errors.New("key required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	// Retain a monotonic tombstone version so an old CAS token cannot be reused.
	s.versions[key]++
	return s.flushLocked()
}

func (s *DurableStore) flushLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	file := durableFile{Values: cloneValues(s.values), Versions: cloneVersions(s.versions)}
	data, err := json.Marshal(file)
	if err != nil {
		return err
	}
	tmp := filepath.Clean(s.path + ".tmp")
	// #nosec G304 -- s.path is the normalized application-selected durable-store path; tmp only appends a fixed suffix.
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(s.path))
	if err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func cloneValues(in map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, len(in))
	for k, v := range in {
		out[k] = append([]byte(nil), v...)
	}
	return out
}

func cloneVersions(in map[string]uint64) map[string]uint64 {
	out := make(map[string]uint64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
