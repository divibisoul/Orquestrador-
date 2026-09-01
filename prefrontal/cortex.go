package prefrontal

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Candidate struct { ID string; Score float64; Cost float64; Risk float64; Utility float64 }
type Decision struct { ID string; Score float64; Reason string; Committed bool; At time.Time }
type Cortex struct { mu sync.RWMutex; decisions []Decision; threshold float64; capacity int }

func New(threshold float64, capacity int) (*Cortex,error) {
	if threshold < 0 || threshold > 1 { return nil,errors.New("threshold must be between 0 and 1") }
	if capacity < 1 { return nil,errors.New("capacity must be positive") }
	return &Cortex{threshold:threshold,capacity:capacity},nil
}

func (c *Cortex) Evaluate(candidates []Candidate) (Candidate,error) {
	if len(candidates)==0{return Candidate{},errors.New("no candidates")}
	best:=candidates[0]
	for _,v:=range candidates { if v.ID==""{continue}; score:=v.Utility-v.Cost-v.Risk; bestScore:=best.Utility-best.Cost-best.Risk; if score>bestScore{best=v} }
	if best.Utility-best.Cost-best.Risk<c.threshold{return Candidate{},errors.New("no candidate exceeds decision threshold")}; return best,nil
}

func (c *Cortex) Plan(candidates []Candidate) ([]Candidate,error) {
	if len(candidates)==0{return nil,errors.New("no candidates")}
	out:=append([]Candidate(nil),candidates...)
	sort.SliceStable(out,func(i,j int)bool{return out[i].Utility-out[i].Cost-out[i].Risk > out[j].Utility-out[j].Cost-out[j].Risk})
	if len(out)>c.capacity{out=out[:c.capacity]}; return out,nil
}

func (c *Cortex) Prioritize(candidates []Candidate) ([]Candidate,error) { return c.Plan(candidates) }

func (c *Cortex) Inhibit(candidate Candidate) bool { return candidate.Risk > candidate.Utility || candidate.Risk >= 1 }

func (c *Cortex) Select(candidates []Candidate) (Candidate,error) {
	planned,err:=c.Plan(candidates); if err!=nil{return Candidate{},err}; for _,v:=range planned{if !c.Inhibit(v){return v,nil}}; return Candidate{},errors.New("all candidates inhibited")
}

func (c *Cortex) ValidateAction(candidate Candidate) error {
	if candidate.ID==""{return errors.New("candidate id is required")}
	if candidate.Cost<0 || candidate.Risk<0{return errors.New("cost and risk cannot be negative")}
	if c.Inhibit(candidate){return errors.New("action inhibited by risk policy")}; return nil
}

func (c *Cortex) Commit(candidate Candidate, reason string) (Decision,error) {
	if err:=c.ValidateAction(candidate);err!=nil{return Decision{},err}
	d:=Decision{ID:candidate.ID,Score:candidate.Utility-candidate.Cost-candidate.Risk,Reason:reason,Committed:true,At:time.Now().UTC()}
	c.mu.Lock(); defer c.mu.Unlock(); c.decisions=append(c.decisions,d); if len(c.decisions)>c.capacity{c.decisions=c.decisions[len(c.decisions)-c.capacity:]}; return d,nil
}

func (c *Cortex) Recall(limit int) []Decision {
	c.mu.RLock(); defer c.mu.RUnlock(); if limit<=0||limit>len(c.decisions){limit=len(c.decisions)}; out:=make([]Decision,limit); copy(out,c.decisions[len(c.decisions)-limit:]); return out
}

func (c *Cortex) Health() map[string]any { c.mu.RLock(); defer c.mu.RUnlock(); return map[string]any{"status":"ready","threshold":c.threshold,"capacity":c.capacity,"decisions":len(c.decisions)} }
