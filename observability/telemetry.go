package observability

import("encoding/json";"log";"net/http";"sort";"sync";"sync/atomic";"time")

type Metrics struct{Requests atomic.Uint64; Success atomic.Uint64; Errors atomic.Uint64; Cancelled atomic.Uint64; InFlight atomic.Int64; LatencyMu sync.Mutex; Latencies []time.Duration}
func(m *Metrics)Observe(start time.Time,err error,cancelled bool){d:=time.Since(start);m.LatencyMu.Lock();if len(m.Latencies)>=2048{copy(m.Latencies,m.Latencies[1:]);m.Latencies=m.Latencies[:2047]};m.Latencies=append(m.Latencies,d);m.LatencyMu.Unlock();if err!=nil{m.Errors.Add(1)}else{m.Success.Add(1)};if cancelled{m.Cancelled.Add(1)}}
func(m *Metrics)Snapshot()map[string]any{m.LatencyMu.Lock();vals:=append([]time.Duration(nil),m.Latencies...);m.LatencyMu.Unlock();sort.Slice(vals,func(i,j int)bool{return vals[i]<vals[j]});p95:=time.Duration(0);if len(vals)>0{idx:=(95*len(vals)+99)/100-1;if idx<0{idx=0};if idx>=len(vals){idx=len(vals)-1};p95=vals[idx]};return map[string]any{"requests":m.Requests.Load(),"success":m.Success.Load(),"errors":m.Errors.Load(),"cancelled":m.Cancelled.Load(),"in_flight":m.InFlight.Load(),"latency_p95_ms":float64(p95)/float64(time.Millisecond)}}

type Logger struct{base map[string]any}
func NewLogger(component string)*Logger{return &Logger{base:map[string]any{"component":component}}}
func(l *Logger)Info(operation,trace string,fields map[string]any){l.emit("INFO",operation,trace,fields)}
func(l *Logger)Error(operation,trace string,err error,fields map[string]any){if fields==nil{fields=map[string]any{}};fields["error"]=err.Error();l.emit("ERROR",operation,trace,fields)}
func(l *Logger)emit(level,operation,trace string,fields map[string]any){v:=map[string]any{"ts":time.Now().UTC().Format(time.RFC3339Nano),"level":level,"operation":operation,"trace_id":trace};for k,x:=range l.base{v[k]=x};for k,x:=range fields{v[k]=x};b,_:=json.Marshal(v);log.Print(string(b))}

func Prometheus(m *Metrics,next http.Handler) http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/metrics"{next.ServeHTTP(w,r);return};s:=m.Snapshot();w.Header().Set("Content-Type","text/plain; version=0.0.4");for _,x:=range []struct{n string;v any}{{"n07_requests_total",s["requests"]},{"n07_success_total",s["success"]},{"n07_errors_total",s["errors"]},{"n07_cancelled_total",s["cancelled"]},{"n07_in_flight",s["in_flight"]},{"n07_latency_p95_ms",s["latency_p95_ms"]}}{_,_=w.Write([]byte(x.n+" "+fmtAny(x.v)+"\n"))}})}
func fmtAny(v any)string{b,_:=json.Marshal(v);return string(b)}
