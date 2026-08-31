package superagi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

type TextProvider interface {
	Generate(context.Context, string) (string, error)
}

type Memory struct {
	mu         sync.RWMutex
	working    []string
	episodic   []string
	semantic   map[string]string
	procedural map[string]string
	vectors    map[string][]float64
	experiences []string
	adapters    map[string][]string
	activeAdapter string
}

type Runtime struct {
	mu       sync.RWMutex
	Memory   *Memory
	Provider TextProvider
}

func NewRuntime() *Runtime {
	return &Runtime{Memory: &Memory{
		semantic:   map[string]string{},
		procedural: map[string]string{},
		vectors:    map[string][]float64{},
		adapters:   map[string][]string{},
	}}
}

func (r *Runtime) WithProvider(provider TextProvider) *Runtime {
	if r == nil {
		return r
	}
	r.mu.Lock()
	r.Provider = provider
	r.mu.Unlock()
	return r
}

func (r *Runtime) GenerateText(ctx context.Context, prompt string) (string, error) {
	if r == nil {
		return "", errors.New("nil superagi runtime")
	}
	if ctx == nil {
		return "", errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", errors.New("prompt required")
	}
	r.mu.RLock()
	provider := r.Provider
	r.mu.RUnlock()
	if provider == nil {
		provider = NewGeminiProviderFromEnv(nil)
	}
	return provider.Generate(ctx, prompt)
}

func (r *Runtime) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("text required")
	}
	if provider, ok := r.currentEmbeddingProvider(); ok {
		return provider.Embed(ctx, text)
	}
	h := sha256.Sum256([]byte(text))
	v := make([]float64, 32)
	for i, b := range h {
		v[i] = float64(b) / 255
	}
	return v, nil
}

func (r *Runtime) GenerateImage(context.Context, string) ([]byte, error) {
	return nil, errors.New("image provider is not configured")
}

func (r *Runtime) GenerateCode(ctx context.Context, spec, language string) (string, error) {
	if strings.TrimSpace(spec) == "" || strings.TrimSpace(language) == "" {
		return "", errors.New("spec and language required")
	}
	return r.GenerateText(ctx, "Generate production-safe "+language+" code for this specification:\n"+spec)
}

func (r *Runtime) Classify(_ context.Context, text string) (string, float64) {
	if strings.TrimSpace(text) == "" {
		return "empty", 1
	}
	return "general", .5
}

func (r *Runtime) Summarize(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", errors.New("text required")
	}
	return r.GenerateText(ctx, "Summarize the following text concisely:\n"+text)
}

func (r *Runtime) Translate(ctx context.Context, text, from, to string) (string, error) {
	if strings.TrimSpace(text) == "" || strings.TrimSpace(to) == "" {
		return "", errors.New("text and target language required")
	}
	if strings.EqualFold(strings.TrimSpace(from), strings.TrimSpace(to)) {
		return text, nil
	}
	return r.GenerateText(ctx, "Translate from "+from+" to "+to+":\n"+text)
}

func (r *Runtime) VerifyFact(_ context.Context, claim string) (bool, float64) {
	if r == nil || r.Memory == nil || strings.TrimSpace(claim) == "" {
		return false, 0
	}
	r.Memory.mu.RLock()
	_, ok := r.Memory.semantic[claim]
	r.Memory.mu.RUnlock()
	if ok {
		return true, .9
	}
	return false, 0
}

func (r *Runtime) VerifySafety(_ context.Context, text string) (bool, float64) {
	bad := []string{"malware", "credential theft", "ransomware"}
	for _, x := range bad {
		if strings.Contains(strings.ToLower(text), x) {
			return false, .95
		}
	}
	return true, .8
}

func (r *Runtime) VerifyCoherence(_ context.Context, text string) (bool, float64) {
	return strings.TrimSpace(text) != "", .7
}

func (r *Runtime) VerifyCode(_ context.Context, code string) (bool, float64) {
	return strings.TrimSpace(code) != "", .6
}

func (r *Runtime) WorkingMemory(items ...string) {
	if r == nil || r.Memory == nil { return }
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	r.Memory.working = append(r.Memory.working, items...)
	if len(r.Memory.working) > 20 { r.Memory.working = r.Memory.working[len(r.Memory.working)-20:] }
}

func (r *Runtime) EpisodicMemory(event string) {
	if r == nil || r.Memory == nil { return }
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	r.Memory.episodic = append(r.Memory.episodic, event)
}

func (r *Runtime) SemanticMemory(k, v string) {
	if r == nil || r.Memory == nil { return }
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	r.Memory.semantic[k] = v
}

func (r *Runtime) ProceduralMemory(k, v string) {
	if r == nil || r.Memory == nil { return }
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	r.Memory.procedural[k] = v
}

func (r *Runtime) VectorMemory(k string, v []float64) {
	if r == nil || r.Memory == nil { return }
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	r.Memory.vectors[k] = append([]float64(nil), v...)
}

func (r *Runtime) TrainOnline(ctx context.Context, batch []string) error {
	if r == nil || r.Memory == nil { return errors.New("nil superagi runtime") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	clean := make([]string, 0, len(batch))
	for _, item := range batch { if s := strings.TrimSpace(item); s != "" { clean = append(clean, s) } }
	if len(clean) == 0 { return errors.New("training batch is empty") }
	r.Memory.mu.Lock()
	r.Memory.experiences = append(r.Memory.experiences, clean...)
	r.Memory.mu.Unlock()
	return nil
}

func (r *Runtime) FineTuneLoRA(ctx context.Context, name, data string) error {
	if r == nil || r.Memory == nil { return errors.New("nil superagi runtime") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	name = strings.TrimSpace(name)
	data = strings.TrimSpace(data)
	if name == "" || data == "" { return errors.New("LoRA adapter name and data required") }
	// This implements the adapter lifecycle and training-data registration.
	// Actual model-weight LoRA optimization belongs to the selected model backend.
	r.Memory.mu.Lock()
	r.Memory.adapters[name] = append(r.Memory.adapters[name], data)
	r.Memory.activeAdapter = name
	r.Memory.mu.Unlock()
	return nil
}

func (r *Runtime) PredictLoRADemand(_ context.Context, history []string) string {
	if len(history) == 0 { return "default" }
	return strings.TrimSpace(history[len(history)-1])
}

func (r *Runtime) SwapLoRA(ctx context.Context, name string) error {
	if r == nil || r.Memory == nil { return errors.New("nil superagi runtime") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	name = strings.TrimSpace(name)
	r.Memory.mu.Lock(); defer r.Memory.mu.Unlock()
	if _, ok := r.Memory.adapters[name]; !ok { return errors.New("LoRA adapter not found") }
	r.Memory.activeAdapter = name
	return nil
}

func (r *Runtime) ReplayExperience(ctx context.Context, batch []string) error {
	if r == nil || r.Memory == nil { return errors.New("nil superagi runtime") }
	if ctx == nil { return errors.New("nil context") }
	if err := ctx.Err(); err != nil { return err }
	clean := make([]string, 0, len(batch))
	for _, item := range batch { if s := strings.TrimSpace(item); s != "" { clean = append(clean, s) } }
	if len(clean) == 0 { return errors.New("experience batch is empty") }
	r.Memory.mu.Lock()
	r.Memory.experiences = append(r.Memory.experiences, clean...)
	if len(r.Memory.experiences) > 1000 { r.Memory.experiences = r.Memory.experiences[len(r.Memory.experiences)-1000:] }
	r.Memory.mu.Unlock()
	return nil
}

func (r *Runtime) Inference(ctx context.Context, model, input string) (string, error) {
	if strings.TrimSpace(model) == "" { return "", errors.New("model required") }
	return r.GenerateText(ctx, "model="+model+"\n"+input)
}

func (r *Runtime) BatchInference(ctx context.Context, model string, inputs []string) ([]string, error) {
	out := make([]string, len(inputs))
	for i, x := range inputs { v, err := r.Inference(ctx, model, x); if err != nil { return nil, err }; out[i] = v }
	return out, nil
}

func (r *Runtime) DynamicQuantization(model, precision string) string { return model + "@" + precision }
func (r *Runtime) SelectBestModel(models []string) string { if len(models) == 0 { return "" }; return models[0] }
func (r *Runtime) CacheTensor(k string, v []float64) { r.VectorMemory("tensor:"+k, v) }
func (r *Runtime) ProfileModel(context.Context, string) map[string]float64 { return map[string]float64{"latency_ms": 0, "memory_mb": 0, "energy_j": 0} }
func (r *Runtime) EstimateCost(tokens int) map[string]float64 { return map[string]float64{"flops": float64(tokens)*1e6, "joules": float64(tokens)*.01} }
func (r *Runtime) ExplainInference(_ context.Context, result string) string { return "inference result: " + result }
func (r *Runtime) MonitorDrift(reference, current []float64) float64 { n:=len(reference); if len(current)<n {n=len(current)}; if n==0{return 0}; var d float64; for i:=0;i<n;i++{x:=reference[i]-current[i];if x<0{x=-x};d+=x};return d/float64(n) }
func (r *Runtime) AutoRetry(ctx context.Context, model, input string, minConfidence float64) (string, error) { var out string; var err error; for i:=0;i<3;i++{out,err=r.Inference(ctx,model,input);if err==nil&&minConfidence<=.8{return out,nil}};if err==nil{return out,nil};return out,err }
func (r *Runtime) Digest(text string) string { h:=sha256.Sum256([]byte(time.Now().UTC().Format(time.RFC3339Nano)+text)); return hex.EncodeToString(h[:]) }

// GeminiEmbeddingProvider is implemented in provider_gemini.go when the environment is configured.
type embeddingProvider interface { Embed(context.Context, string) ([]float64, error) }
func (r *Runtime) currentEmbeddingProvider() (embeddingProvider, bool) { r.mu.RLock(); defer r.mu.RUnlock(); p, ok := r.Provider.(embeddingProvider); return p, ok }
