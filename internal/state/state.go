package state

import("sync";"time")

type Entry struct{Key string;Value []byte;Version uint64;UpdatedAt time.Time}
type Store struct{mu sync.RWMutex;data map[string]Entry;version uint64}
func New()*Store{return &Store{data:map[string]Entry{}}}
func(s *Store) Put(key string,value []byte)Entry{s.mu.Lock();defer s.mu.Unlock();s.version++;e:=Entry{key,append([]byte(nil),value...),s.version,time.Now()};s.data[key]=e;return e}
func(s *Store) Get(key string)(Entry,bool){s.mu.RLock();defer s.mu.RUnlock();e,ok:=s.data[key];if ok{e.Value=append([]byte(nil),e.Value...)};return e,ok}
func(s *Store) Delete(key string)bool{s.mu.Lock();defer s.mu.Unlock();if _,ok:=s.data[key];!ok{return false};delete(s.data,key);s.version++;return true}
func(s *Store) Snapshot()[]Entry{s.mu.RLock();defer s.mu.RUnlock();out:=make([]Entry,0,len(s.data));for _,e:=range s.data{out=append(out,e)};return out}

// RaftBoundary defines the production consensus integration point without forcing a vendor.
type RaftBoundary interface{ Apply([]byte) error; Leader() bool }
