package superagi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Model struct { Name string; Precision string; Score float64; LatencyMS float64 }

type AGI struct {
	mu sync.RWMutex
	Runtime *Runtime
	Models []Model
	ActiveLoRA string
	lastError error
}

func New() *AGI { return &AGI{Runtime: NewRuntime(), Models: []Model{{Name:"default",Precision:"FP32",Score:.7,LatencyMS:10}}} }
func NewWithProvider(provider TextProvider) *AGI { return New().WithProvider(provider) }
func (a *AGI) WithProvider(provider TextProvider) *AGI { if a==nil{return a};a.mu.Lock();defer a.mu.Unlock();if a.Runtime==nil{a.Runtime=NewRuntime()};a.Runtime.Provider=provider;return a }
func (a *AGI) setError(err error){if a==nil{return};a.mu.Lock();a.lastError=err;a.mu.Unlock()}
func (a *AGI) LastError() error {if a==nil{return errors.New("nil AGI")};a.mu.RLock();defer a.mu.RUnlock();return a.lastError}
func (a *AGI) GenerateText(prompt string)string{if a==nil||a.Runtime==nil{return ""};out,err:=a.Runtime.GenerateText(context.Background(),prompt);a.setError(err);return out}
func (a *AGI) GenerateEmbedding(text string)[]float64{if a==nil||a.Runtime==nil{return nil};out,err:=a.Runtime.GenerateEmbedding(context.Background(),text);a.setError(err);return out}
func (a *AGI) GenerateImage(prompt string)([]byte,error){if a==nil||a.Runtime==nil{return nil,errors.New("nil AGI")};return a.Runtime.GenerateImage(context.Background(),prompt)}
func (a *AGI) GenerateCode(spec,language string)string{if a==nil||a.Runtime==nil{return ""};out,err:=a.Runtime.GenerateCode(context.Background(),spec,language);a.setError(err);return out}
func (a *AGI) Classify(text string,categories []string)(string,float64){if strings.TrimSpace(text)==""{return "empty",1};if len(categories)==0{return "",0};lower:=strings.ToLower(text);for _,c:=range categories{if strings.Contains(lower,strings.ToLower(c)){return c,1}};return categories[0],0}
func (a *AGI) Summarize(text string,maxChars int)string{if a==nil||a.Runtime==nil{return ""};out,err:=a.Runtime.Summarize(context.Background(),text);a.setError(err);if err!=nil{return ""};if maxChars>0&&len(out)>maxChars{return out[:maxChars]};return out}
func (a *AGI) Translate(text,from,to string)(string,error){if a==nil||a.Runtime==nil{return "",errors.New("nil AGI")};return a.Runtime.Translate(context.Background(),text,from,to)}
func (a *AGI) VerifyFact(claim string)bool{if a==nil||a.Runtime==nil{return false};ok,_:=a.Runtime.VerifyFact(context.Background(),claim);return ok}
func (a *AGI) VerifySafety(output string)bool{if a==nil||a.Runtime==nil{return false};ok,_:=a.Runtime.VerifySafety(context.Background(),output);return ok}
func (a *AGI) VerifyCoherence(text string)bool{if a==nil||a.Runtime==nil{return false};ok,_:=a.Runtime.VerifyCoherence(context.Background(),text);return ok}
func (a *AGI) VerifyCode(code,language string)bool{if a==nil||a.Runtime==nil{return false};ok,_:=a.Runtime.VerifyCode(context.Background(),code);return ok}
func (a *AGI) WorkingMemory(items ...string){if a!=nil&&a.Runtime!=nil{a.Runtime.WorkingMemory(items...)}}
func (a *AGI) EpisodicMemory(event string){if a!=nil&&a.Runtime!=nil{a.Runtime.EpisodicMemory(event)}}
func (a *AGI) SemanticMemory(key,value string){if a!=nil&&a.Runtime!=nil{a.Runtime.SemanticMemory(key,value)}}
func (a *AGI) ProceduralMemory(name string,steps []string){if a!=nil&&a.Runtime!=nil{a.Runtime.ProceduralMemory(name,strings.Join(steps,"\n"))}}
func (a *AGI) VectorMemory(key string,v []float64)[]string{if a==nil||a.Runtime==nil{return nil};a.Runtime.VectorMemory(key,v);return []string{key}}
func (a *AGI) TrainOnline(batch []string)int{if a==nil||a.Runtime==nil{return 0};if err:=a.Runtime.TrainOnline(context.Background(),batch);err!=nil{a.setError(err);return 0};return len(batch)}
func (a *AGI) FineTuneLoRA(name string,data []string)error{if a==nil||a.Runtime==nil{return errors.New("nil AGI")};return a.Runtime.FineTuneLoRA(context.Background(),name,strings.Join(data,"\n"))}
func (a *AGI) PredictLoRADemand()string{if a==nil{return "default"};a.mu.RLock();defer a.mu.RUnlock();if a.ActiveLoRA!=""{return a.ActiveLoRA};return "default"}
func (a *AGI) SwapLoRA(name string)error{if a==nil||a.Runtime==nil{return errors.New("nil AGI")};if err:=a.Runtime.SwapLoRA(context.Background(),name);err!=nil{return err};a.mu.Lock();a.ActiveLoRA=name;a.mu.Unlock();return nil}
func (a *AGI) ReplayExperience(n int)[]string{if a==nil||a.Runtime==nil{return nil};return a.Runtime.ExperienceSnapshot(n)}
func (a *AGI) Inference(input string)string{return a.GenerateText(input)}
func (a *AGI) BatchInference(inputs []string)[]string{out:=make([]string,len(inputs));for i,x:=range inputs{out[i]=a.Inference(x)};return out}
func (a *AGI) DynamicQuantization(model Model,precision string)Model{model.Precision=precision;return model}
func (a *AGI) SelectBestModel(maxLatency float64)Model{if a==nil{return Model{}};a.mu.RLock();defer a.mu.RUnlock();best:=Model{};for _,m:=range a.Models{if m.LatencyMS<=maxLatency&&m.Score>best.Score{best=m}};return best}
func (a *AGI) CacheTensor(key string,v []float64){if a!=nil&&a.Runtime!=nil{a.Runtime.CacheTensor(key,v)}}
func (a *AGI) ProfileModel(model Model)map[string]float64{if a==nil||a.Runtime==nil{return nil};return a.Runtime.ProfileModel(context.Background(),model.Name)}
func (a *AGI) EstimateCost(tokens int)map[string]float64{if a==nil||a.Runtime==nil{return nil};return a.Runtime.EstimateCost(tokens)}
func (a *AGI) ExplainInference(input,output string)string{return fmt.Sprintf("input_chars=%d output_chars=%d model=%s",len(input),len(output),a.SelectBestModel(1e9).Name)}
func (a *AGI) MonitorDrift(reference,current []float64)float64{if a==nil||a.Runtime==nil{return 0};return a.Runtime.MonitorDrift(reference,current)}
func (a *AGI) AutoRetry(input string,confidence float64,maxRetries int)string{if a==nil||a.Runtime==nil{return ""};out,err:=a.Runtime.AutoRetry(context.Background(),"default",input,confidence);a.setError(err);return out}
