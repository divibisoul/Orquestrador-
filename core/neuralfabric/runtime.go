package neuralfabric

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

var ErrNilContext = errors.New("nil context")

type Runtime struct {
	mu         sync.RWMutex
	history    []Experience
	weights    [5]float64
	checkpoint []byte
}

func NewRuntime() *Runtime { return &Runtime{weights: [5]float64{.35, .2, .2, .15, .1}} }
func (r *Runtime) EncodeState(ctx context.Context, s State) (Vector, error) { if err:=contextErr(ctx);err!=nil{return nil,err};return r.Normalize(append(Vector{s.Constraints["latency"],s.Constraints["energy"],s.Constraints["quality"]},s.Features...)),nil }
func (r *Runtime) EncodeTask(ctx context.Context, task string) (Vector, error) { if err:=contextErr(ctx);err!=nil{return nil,err};v:=make(Vector,8);for i,b:=range []byte(task){v[i%len(v)]+=float64(b)/255};return v,nil }
func (r *Runtime) Normalize(v Vector) Vector { var n float64;for _,x:=range v{n+=x*x};n=math.Sqrt(n);if n==0{return v};o:=make(Vector,len(v));for i,x:=range v{o[i]=x/n};return o }
func (r *Runtime) Latency(s State,a Route) Prediction {base:=100.0;if a.Precision=="int8"{base*=.55};if a.Precision=="int4"{base*=.4};if a.BatchSize>1{base*=.7};return Prediction{Value:base+float64(len(s.Features))*2,Confidence:.65}}
func (r *Runtime) Cost(s State,a Route) Prediction{return Prediction{Value:r.Latency(s,a).Value*.001,Confidence:.7}}
func (r *Runtime) Energy(s State,a Route) Prediction{return Prediction{Value:r.Latency(s,a).Value*.02,Confidence:.6}}
func (r *Runtime) Quality(_ State,a Route) Prediction{q:=.95;if a.Precision=="int4"{q=.82};return Prediction{Value:q,Confidence:.8}}
func (r *Runtime) Failure(_ State,a Route) Prediction{if a.NodeID==""{return Prediction{Value:1,Confidence:1}};return Prediction{Value:.02,Confidence:.6}}
func (r *Runtime) Route(ctx context.Context,s State,candidates []Route)(Route,error){if err:=contextErr(ctx);err!=nil{return Route{},err};type scored struct{r Route;score float64};x:=make([]scored,0,len(candidates));for _,a:=range candidates{l:=r.Latency(s,a);c:=r.Cost(s,a);e:=r.Energy(s,a);q:=r.Quality(s,a);f:=r.Failure(s,a);score:=.35*l.Value+.2*c.Value+.2*e.Value+2*(1-q.Value)+10*f.Value;a.Score=score;a.Confidence=(l.Confidence+c.Confidence+q.Confidence)/3;x=append(x,scored{a,score})};if len(x)==0{return Route{},errors.New("no route candidates")};sort.Slice(x,func(i,j int)bool{return x[i].score<x[j].score});return x[0].r,nil}
func (r *Runtime) Observe(e Experience){if r==nil{return};r.mu.Lock();defer r.mu.Unlock();r.history=append(r.history,e);if len(r.history)>10000{r.history=append([]Experience(nil),r.history[len(r.history)-10000:]...)}}
func (r *Runtime) Reward(e Experience)float64{return math.Max(-1,math.Min(1,e.Reward))}
func (r *Runtime) Update(ctx context.Context)error{if err:=contextErr(ctx);err!=nil{return err};if r==nil{return errors.New("nil neural fabric runtime")};r.mu.Lock();defer r.mu.Unlock();if len(r.history)==0{return nil};var rewardSum float64;for _,e:=range r.history{rewardSum+=r.Reward(e)};avg:=rewardSum/float64(len(r.history));adj:=math.Max(-.05,math.Min(.05,avg*.02));r.weights[0]=clamp01(r.weights[0]+adj);r.weights[1]=clamp01(r.weights[1]+adj*.5);r.weights[2]=clamp01(r.weights[2]-adj*.5);r.weights[3]=clamp01(r.weights[3]-adj*.25);r.weights[4]=clamp01(r.weights[4]-adj*.25);return nil}
func (r *Runtime) Save(ctx context.Context)error{if err:=contextErr(ctx);err!=nil{return err};if r==nil{return errors.New("nil neural fabric runtime")};r.mu.Lock();defer r.mu.Unlock();snapshot:=struct{History []Experience `json:"history"`;Weights [5]float64 `json:"weights"`}{History:append([]Experience(nil),r.history...),Weights:r.weights};data,err:=json.Marshal(snapshot);if err!=nil{return err};r.checkpoint=append([]byte(nil),data...);return nil}
func (r *Runtime) Load(ctx context.Context)error{if err:=contextErr(ctx);err!=nil{return err};if r==nil{return errors.New("nil neural fabric runtime")};r.mu.Lock();defer r.mu.Unlock();if len(r.checkpoint)==0{return errors.New("neural fabric checkpoint not available")};var snapshot struct{History []Experience `json:"history"`;Weights [5]float64 `json:"weights"`};if err:=json.Unmarshal(r.checkpoint,&snapshot);err!=nil{return err};r.history=append([]Experience(nil),snapshot.History...);r.weights=snapshot.Weights;return nil}
func (r *Runtime) ExperienceCount()int{if r==nil{return 0};r.mu.RLock();defer r.mu.RUnlock();return len(r.history)}
func (r *Runtime) RecordFeedback(e Experience){if r==nil{return};e.Done=true;e.Reward=r.Reward(e);r.Observe(e);_ = r.Update(context.Background())}
func (r *Runtime) Now()time.Time{return time.Now()}
func contextErr(ctx context.Context)error{if ctx==nil{return ErrNilContext};if err:=ctx.Err();err!=nil{return err};return nil}
func clamp01(v float64)float64{return math.Max(0,math.Min(1,v))}
