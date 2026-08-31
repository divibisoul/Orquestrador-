package cache

import("container/list";"sync")
type Cache struct{mu sync.Mutex;cap int;items map[string]*list.Element;lru *list.List};type item struct{k string;v any}
func New(capacity int)*Cache{if capacity<1{capacity=1};return &Cache{cap:capacity,items:map[string]*list.Element{},lru:list.New()}}
func(c *Cache)Get(k string)(any,bool){c.mu.Lock();defer c.mu.Unlock();e,ok:=c.items[k];if !ok{return nil,false};c.lru.MoveToFront(e);return e.Value.(item).v,true}
func(c *Cache)Set(k string,v any){c.mu.Lock();defer c.mu.Unlock();if e,ok:=c.items[k];ok{e.Value=item{k,v};c.lru.MoveToFront(e);return};c.items[k]=c.lru.PushFront(item{k,v});for len(c.items)>c.cap{e:=c.lru.Back();delete(c.items,e.Value.(item).k);c.lru.Remove(e)}}
