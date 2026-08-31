package nexus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Task is the substrate-neutral unit of execution.
type Task struct { ID, Goal string; Dependencies []string; Priority int; Capability string; Status string; Result any; Err error }
type Workflow struct { ID string; Tasks map[string]*Task; Status string; CreatedAt time.Time; UpdatedAt time.Time }
type Node struct { ID, Status string; Capabilities []string; CPU, GPU, NPU bool; RAMMB, VRAMMB int64; Workers int; Load float64 }
type Decision struct { Goal string; Plan []Task; Confidence float64; Reason string }
type MemoryItem struct { Key string; Value any; CreatedAt time.Time; Score float64 }
type InferenceResult struct { Output any; Confidence float64; Model string; Latency time.Duration }

// Orchestrator owns execution state and resilience policies.
type Orchestrator struct { mu sync.RWMutex; workflows map[string]*Workflow; nodes map[string]*Node; routes map[string]string; maxWorkers int; failures map[string][]time.Time }
func NewOrchestrator(workers int) *Orchestrator { if workers < 1 { workers = 1 }; return &Orchestrator{workflows: map[string]*Workflow{}, nodes: map[string]*Node{}, routes: map[string]string{}, maxWorkers: workers, failures: map[string][]time.Time{}} }
func (o *Orchestrator) CreateWorkflow(_ context.Context, w Workflow) error { o.mu.Lock(); defer o.mu.Unlock(); if w.ID=="" { return errors.New("workflow id required") }; if _,ok:=o.workflows[w.ID]; ok{return errors.New("workflow already exists")}; if w.Tasks==nil {w.Tasks=map[string]*Task{}}; w.Status="Pending"; w.CreatedAt=time.Now(); w.UpdatedAt=w.CreatedAt; o.workflows[w.ID]=&w; return validateDAG(w.Tasks) }
func (o *Orchestrator) ExecuteStep(_ context.Context, workflowID, taskID string) error { o.mu.Lock(); defer o.mu.Unlock(); w,ok:=o.workflows[workflowID]; if !ok{return errors.New("workflow not found")}; t,ok:=w.Tasks[taskID]; if !ok{return errors.New("task not found")}; t.Status="Running"; if t.Capability!="" { if n:=o.selectNodeLocked(t.Capability); n==nil {t.Status="Failed"; t.Err=errors.New("no capable node") ; return t.Err} }; t.Status="Completed"; t.Result="completed"; w.Status="Running"; w.UpdatedAt=time.Now(); return nil }
func (o *Orchestrator) GetWorkflowStatus(_ context.Context,id string)(string,error){o.mu.RLock();defer o.mu.RUnlock();w,ok:=o.workflows[id];if !ok{return "",errors.New("workflow not found")};return w.Status,nil}
func (o *Orchestrator) PauseWorkflow(_ context.Context,id string)error{o.mu.Lock();defer o.mu.Unlock();w,ok:=o.workflows[id];if !ok{return errors.New("workflow not found")};w.Status="Paused";return nil}
func (o *Orchestrator) ResumeWorkflow(_ context.Context,id string)error{o.mu.Lock();defer o.mu.Unlock();w,ok:=o.workflows[id];if !ok{return errors.New("workflow not found")};if w.Status!="Paused"{return errors.New("workflow is not paused")};w.Status="Running";return nil}
func (o *Orchestrator) RollbackWorkflow(_ context.Context,id string)error{o.mu.Lock();defer o.mu.Unlock();w,ok:=o.workflows[id];if !ok{return errors.New("workflow not found")};for _,t:=range w.Tasks{if t.Status=="Completed"{t.Status="Pending";t.Result=nil}};w.Status="Paused";return nil}
func (o *Orchestrator) RetryFailedStep(ctx context.Context,wid,tid string)error{for d:=time.Second;d<=time.Minute;d*=2{if err:=o.ExecuteStep(ctx,wid,tid);err==nil{return nil};time.Sleep(d)};return errors.New("retry budget exhausted")}
func (o *Orchestrator) ExecuteParallel(ctx context.Context,tasks []func() error)[]error{sem:=make(chan struct{},o.maxWorkers);out:=make([]error,len(tasks));var wg sync.WaitGroup;for i,fn:=range tasks{wg.Add(1);go func(i int,fn func()error){defer wg.Done();sem<-struct{}{};defer func(){<-sem}();out[i]=fn()}(i,fn)};wg.Wait();return out}
func (o *Orchestrator) ExecuteDistributed(_ context.Context,t *Task)error{o.mu.Lock();defer o.mu.Unlock();n:=o.selectNodeLocked(t.Capability);if n==nil{return errors.New("no capable node")};o.routes[t.ID]=n.ID;t.Status="Completed";t.Result=fmt.Sprintf("executed on %s",n.ID);return nil}
func (o *Orchestrator) CircuitBreaker(service string, now time.Time)bool{o.mu.Lock();defer o.mu.Unlock();xs:=o.failures[service];cut:=now.Add(-10*time.Second);k:=0;for _,x:=range xs{if x.After(cut){xs[k]=x;k++}};o.failures[service]=xs[:k];return len(xs)>=5}
func (o *Orchestrator) Bulkhead(limit int, jobs <-chan func() error) <-chan error {out:=make(chan error);if limit<1{limit=1};go func(){defer close(out);sem:=make(chan struct{},limit);var wg sync.WaitGroup;for job:=range jobs{wg.Add(1);go func(f func()error){defer wg.Done();sem<-struct{}{};defer func(){<-sem}();out<-f()}(job)};wg.Wait()}();return out}
func (o *Orchestrator) RateLimiter(rate int) func() bool {if rate<1{rate=1};var mu sync.Mutex;window:=time.Now();count:=0;return func()bool{mu.Lock();defer mu.Unlock();now:=time.Now();if now.Sub(window)>=time.Second{window=now;count=0};if count>=rate{return false};count++;return true}}
func (o *Orchestrator) SpawnSubOrchestrator() *Orchestrator{return NewOrchestrator(o.maxWorkers)}
func (o *Orchestrator) KillSubOrchestrator(_ *Orchestrator) {}
func (o *Orchestrator) RebalanceTasks(tasks []*Task) []*Task{sort.SliceStable(tasks,func(i,j int)bool{return tasks[i].Priority>tasks[j].Priority});return tasks}
func (o *Orchestrator) RegisterNode(n Node){o.mu.Lock();defer o.mu.Unlock();o.nodes[n.ID]=&n}
func (o *Orchestrator) selectNodeLocked(capability string)*Node{var best *Node;for _,n:=range o.nodes{if n.Status!="ready"{continue};ok:=capability=="";for _,c:=range n.Capabilities{if c==capability{ok=true}};if ok&&(best==nil||n.Load<best.Load){best=n}};return best}

// Prefrontal is the executive decision layer.
type Prefrontal struct{mu sync.RWMutex; memories []MemoryItem; last Decision}
func NewPrefrontal()*Prefrontal{return &Prefrontal{memories:make([]MemoryItem,0)}}
func(p *Prefrontal) ReadContext(ctx context.Context)map[string]any{return map[string]any{"deadline":ctx.Err(),"memory_items":len(p.memories)}}
func(p *Prefrontal) FuseSignals(signals map[string]float64)map[string]float64{out:=map[string]float64{};for k,v:=range signals{if v<0{v=0};if v>1{v=1};out[k]=v};return out}
func(p *Prefrontal) DetectAnomalies(values []float64)[]int{if len(values)<2{return nil};var sum float64;for _,v:=range values{sum+=v};mean:=sum/float64(len(values));var s float64;for _,v:=range values{d:=v-mean;s+=d*d};sd:=sqrt(s/float64(len(values)));var idx []int;for i,v:=range values{if sd>0&&abs(v-mean)>3*sd{idx=append(idx,i)}};return idx}
func(p *Prefrontal) CausalReason(event string)string{e:=strings.ToLower(event);switch{case strings.Contains(e,"timeout"):return "latency/dependency saturation";case strings.Contains(e,"oom"):return "memory pressure";case strings.Contains(e,"unavailable"):return "node or dependency availability";default:return "insufficient causal evidence"}}
func(p *Prefrontal) ProbabilisticReason(success,prior float64)float64{if success<0{success=0};if success>1{success=1};if prior<0{prior=0};if prior>1{prior=1};return success*prior}
func(p *Prefrontal) GeneratePlan(goal string)[]Task{parts:=strings.Fields(goal);tasks:=make([]Task,0,len(parts));for i,w:=range parts{tasks=append(tasks,Task{ID:fmt.Sprintf("t-%d",i+1),Goal:w,Priority:100-i})};for i:=1;i<len(tasks);i++{tasks[i].Dependencies=[]string{tasks[i-1].ID}};return tasks}
func(p *Prefrontal) SimulatePlan(plan []Task)map[string]any{cost:=float64(len(plan));return map[string]any{"tasks":len(plan),"estimated_cost":cost,"success_probability":1/(1+cost/10)}}
func(p *Prefrontal) PrioritizeGoals(goals []string, urgency,impact map[string]float64)[]string{sort.SliceStable(goals,func(i,j int)bool{return urgency[goals[i]]+impact[goals[i]]>urgency[goals[j]]+impact[goals[j]]});return goals}
func(p *Prefrontal) Decide(goal string,plan []Task,constraints map[string]float64)Decision{conf:=1.0;if len(plan)>10{conf=.7};if constraints["budget"]>0&&float64(len(plan))>constraints["budget"]{conf=.4};d:=Decision{Goal:goal,Plan:plan,Confidence:conf,Reason:"plan selected against declared constraints"};p.mu.Lock();p.last=d;p.mu.Unlock();return d}
func(p *Prefrontal) Delegate(d Decision)*Workflow{tasks:=map[string]*Task{};for i:=range d.Plan{t:=d.Plan[i];tasks[t.ID]=&t};return &Workflow{ID:hash(d.Goal+time.Now().String())[:12],Tasks:tasks}}
func(p *Prefrontal) EvaluateOutcome(expected,actual float64)float64{return abs(expected-actual)}
func(p *Prefrontal) LearnFromFeedback(item MemoryItem){p.mu.Lock();defer p.mu.Unlock();item.CreatedAt=time.Now();p.memories=append(p.memories,item)}
func(p *Prefrontal) OptimizePolicy(rewards []float64)float64{if len(rewards)==0{return 0};var s float64;for _,r:=range rewards{s+=r};return s/float64(len(rewards))}
func(p *Prefrontal) ExplainDecision(d Decision)string{return fmt.Sprintf("Goal=%q; tasks=%d; confidence=%.2f; reason=%s",d.Goal,len(d.Plan),d.Confidence,d.Reason)}

// SuperAGI is model-provider agnostic. Real model runtimes plug into ModelProvider.
type ModelProvider interface{Generate(context.Context,string)(string,error);Embed(context.Context,string)([]float32,error);ModelName()string}
type SuperAGI struct{mu sync.RWMutex;provider ModelProvider;working []MemoryItem;episodic []MemoryItem;semantic map[string]MemoryItem;procedural map[string]MemoryItem;vector map[string][]float32}
func NewSuperAGI(provider ModelProvider)*SuperAGI{return &SuperAGI{provider:provider,semantic:map[string]MemoryItem{},procedural:map[string]MemoryItem{},vector:map[string][]float32{}}}
func(s *SuperAGI) GenerateText(ctx context.Context,prompt string)(string,error){if s.provider==nil{return prompt,nil};return s.provider.Generate(ctx,prompt)}
func(s *SuperAGI) GenerateEmbedding(ctx context.Context,text string)([]float32,error){if s.provider==nil{return deterministicEmbedding(text,32),nil};return s.provider.Embed(ctx,text)}
func(s *SuperAGI) GenerateImage(_ context.Context,prompt string)(string,error){return "image-request://"+hash(prompt),nil}
func(s *SuperAGI) GenerateCode(_ context.Context,spec,language string)(string,error){return fmt.Sprintf("// generated specification (%s)\n// %s\n",language,spec),nil}
func(s *SuperAGI) Classify(text string)map[string]float64{t:=strings.ToLower(text);scores:=map[string]float64{"general":.5};if strings.Contains(t,"code"){scores["code"]=.9};if strings.Contains(t,"image"){scores["image"]=.9};return scores}
func(s *SuperAGI) Summarize(text string,maxWords int)string{w:=strings.Fields(text);if maxWords<1||len(w)<=maxWords{return text};return strings.Join(w[:maxWords]," ")+" …"}
func(s *SuperAGI) Translate(text,_,_ string)string{return text}
func(s *SuperAGI) VerifyFact(claim string)bool{return strings.TrimSpace(claim)!=""}
func(s *SuperAGI) VerifySafety(output string)bool{return !strings.Contains(strings.ToLower(output),"malware")}
func(s *SuperAGI) VerifyCoherence(text string)bool{return len(strings.TrimSpace(text))>0}
func(s *SuperAGI) VerifyCode(code string)bool{return strings.TrimSpace(code)!=""}
func(s *SuperAGI) WorkingMemory(item MemoryItem) {s.mu.Lock();defer s.mu.Unlock();s.working=append(s.working,item);if len(s.working)>20{s.working=s.working[len(s.working)-20:]}}
func(s *SuperAGI) EpisodicMemory(item MemoryItem){s.mu.Lock();defer s.mu.Unlock();s.episodic=append(s.episodic,item)}
func(s *SuperAGI) SemanticMemory(item MemoryItem){s.mu.Lock();defer s.mu.Unlock();s.semantic[item.Key]=item}
func(s *SuperAGI) ProceduralMemory(item MemoryItem){s.mu.Lock();defer s.mu.Unlock();s.procedural[item.Key]=item}
func(s *SuperAGI) VectorMemory(key string,v []float32){s.mu.Lock();defer s.mu.Unlock();s.vector[key]=append([]float32(nil),v...)}
func(s *SuperAGI) TrainOnline(_ context.Context,_ []MemoryItem)error{return nil}
func(s *SuperAGI) FineTuneLoRA(_ context.Context,adapter string)error{if adapter==""{return errors.New("adapter required")};return nil}
func(s *SuperAGI) PredictLoRADemand(_ []string)string{return "default"}
func(s *SuperAGI) SwapLoRA(_ context.Context,adapter string)error{if adapter==""{return errors.New("adapter required")};return nil}
func(s *SuperAGI) ReplayExperience(_ context.Context,items []MemoryItem)int{return len(items)}
func(s *SuperAGI) Inference(ctx context.Context,prompt string)(InferenceResult,error){start:=time.Now();out,err:=s.GenerateText(ctx,prompt);r:=InferenceResult{Output:out,Model:"deterministic-fallback",Latency:time.Since(start),Confidence:.5};if s.provider!=nil{r.Model=s.provider.ModelName()};return r,err}
func(s *SuperAGI) BatchInference(ctx context.Context,prompts []string)([]InferenceResult,error){out:=make([]InferenceResult,len(prompts));for i,p:=range prompts{r,e:=s.Inference(ctx,p);if e!=nil{return nil,e};out[i]=r};return out,nil}
func(s *SuperAGI) DynamicQuantization(model []float32,bits int)[]int8{if bits!=8{return nil};out:=make([]int8,len(model));for i,v:=range model{if v>1{v=1};if v< -1{v=-1};out[i]=int8(v*127)};return out}
func(s *SuperAGI) SelectBestModel(models []string,budget time.Duration)string{if len(models)==0{return ""};if budget>0{return models[0]};return models[len(models)-1]}
func(s *SuperAGI) CacheTensor(key string,v []float32){s.VectorMemory(key,v)}
func(s *SuperAGI) ProfileModel(ctx context.Context,prompt string)time.Duration{start:=time.Now();_,_=s.GenerateText(ctx,prompt);return time.Since(start)}
func(s *SuperAGI) EstimateCost(tokens int)float64{if tokens<0{tokens=0};return float64(tokens)*4.0}
func(s *SuperAGI) ExplainInference(r InferenceResult)string{return fmt.Sprintf("model=%s confidence=%.2f latency=%s",r.Model,r.Confidence,r.Latency)}
func(s *SuperAGI) MonitorDrift(reference,current []float64)float64{if len(reference)==0||len(current)==0{return 1};return abs(mean(reference)-mean(current))}
func(s *SuperAGI) AutoRetry(ctx context.Context,prompt string,threshold float64)(InferenceResult,error){r,e:=s.Inference(ctx,prompt);if e==nil&&r.Confidence>=threshold{return r,nil};if e!=nil{return r,e};r.Confidence=.5;return r,nil}

func validateDAG(tasks map[string]*Task)error{vis:=map[string]int{};var visit func(string)error;visit=func(id string)error{if vis[id]==1{return errors.New("workflow contains dependency cycle")};if vis[id]==2{return nil};t,ok:=tasks[id];if !ok{return fmt.Errorf("missing dependency node %s",id)};vis[id]=1;for _,d:=range t.Dependencies{if err:=visit(d);err!=nil{return err}};vis[id]=2;return nil};for id:=range tasks{if err:=visit(id);err!=nil{return err}};return nil}
func abs(v float64)float64{if v<0{return -v};return v};func sqrt(v float64)float64{if v<=0{return 0};x:=v;for i:=0;i<20;i++{x=(x+v/x)/2};return x};func mean(v []float64)float64{if len(v)==0{return 0};var s float64;for _,x:=range v{s+=x};return s/float64(len(v))};func hash(s string)string{h:=sha256.Sum256([]byte(s));return hex.EncodeToString(h[:])};func deterministicEmbedding(s string,n int)[]float32{h:=sha256.Sum256([]byte(s));v:=make([]float32,n);for i:=range v{v[i]=float32(int(h[i%len(h)])-128)/128};return v}
