package state

import("sync";"time")

type Store struct{mu sync.RWMutex; values map[string][]byte; versions map[string]uint64}
func NewStore()*Store{return &Store{values:map[string][]byte{},versions:map[string]uint64{}}}
func(s *Store)Put(key string,value []byte)uint64{s.mu.Lock();defer s.mu.Unlock();s.values[key]=append([]byte(nil),value...);s.versions[key]++;return s.versions[key]}
func(s *Store)Get(key string)([]byte,uint64,bool){s.mu.RLock();defer s.mu.RUnlock();v,ok:=s.values[key];return append([]byte(nil),v...),s.versions[key],ok}
func(s *Store)Delete(key string){s.mu.Lock();defer s.mu.Unlock();delete(s.values,key);delete(s.versions,key)}
func(s *Store)CompareAndSwap(key string,version uint64,value []byte)bool{s.mu.Lock();defer s.mu.Unlock();if s.versions[key]!=version{return false};s.values[key]=append([]byte(nil),value...);s.versions[key]++;return true}
func(s *Store)Checkpoint()map[string]any{s.mu.RLock();defer s.mu.RUnlock();return map[string]any{"keys":len(s.values),"timestamp":time.Now().Unix()}}
