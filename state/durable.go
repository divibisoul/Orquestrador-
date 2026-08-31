package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type DurableStore struct { mu sync.RWMutex; path string; values map[string][]byte; versions map[string]uint64 }
type durableFile struct { Values map[string][]byte `json:"values"`; Versions map[string]uint64 `json:"versions"` }

func OpenDurable(path string) (*DurableStore,error) {
	if path==""{return nil,errors.New("state path required")}
	s:=&DurableStore{path:path,values:map[string][]byte{},versions:map[string]uint64{}}
	data,err:=os.ReadFile(path);if os.IsNotExist(err){return s,nil};if err!=nil{return nil,err}
	var file durableFile;if err:=json.Unmarshal(data,&file);err!=nil{return nil,err};if file.Values!=nil{s.values=cloneValues(file.Values)};if file.Versions!=nil{s.versions=file.Versions};return s,nil
}
func(s *DurableStore)Put(key string,value []byte)(uint64,error){if s==nil{return 0,errors.New("nil durable store")};if key==""{return 0,errors.New("key required")};s.mu.Lock();defer s.mu.Unlock();s.values[key]=append([]byte(nil),value...);s.versions[key]++;if err:=s.flushLocked();err!=nil{return 0,err};return s.versions[key],nil}
func(s *DurableStore)Get(key string)([]byte,uint64,bool){if s==nil{return nil,0,false};s.mu.RLock();defer s.mu.RUnlock();v,ok:=s.values[key];return append([]byte(nil),v...),s.versions[key],ok}
func(s *DurableStore)Delete(key string)error{if s==nil{return errors.New("nil durable store")};s.mu.Lock();defer s.mu.Unlock();delete(s.values,key);delete(s.versions,key);return s.flushLocked()}
func(s *DurableStore)flushLocked()error{file:=durableFile{Values:cloneValues(s.values),Versions:map[string]uint64{}};for k,v:=range s.versions{file.Versions[k]=v};data,err:=json.Marshal(file);if err!=nil{return err};if dir:=filepath.Dir(s.path);dir!="."{if err:=os.MkdirAll(dir,0700);err!=nil{return err}};tmp:=s.path+".tmp";if err:=os.WriteFile(tmp,data,0600);err!=nil{return err};return os.Rename(tmp,s.path)}
func cloneValues(in map[string][]byte)map[string][]byte{out:=make(map[string][]byte,len(in));for k,v:=range in{out[k]=append([]byte(nil),v...)};return out}
