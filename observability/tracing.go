package observability

import("context";"crypto/rand";"encoding/hex";"errors";"sync";"time")
type traceKey struct{}
type Span struct{mu sync.Mutex;operation string;traceID string;spanID string;start time.Time;end time.Time;err string;ended bool}
type TraceRecord struct{TraceID string;SpanID string;Operation string;StartedAt time.Time;EndedAt time.Time;Duration time.Duration;Error string}
type recorder struct{mu sync.RWMutex;records []TraceRecord;capacity int}
var globalRecorder=&recorder{capacity:4096}
func newID(n int)(string,error){b:=make([]byte,n);if _,err:=rand.Read(b);err!=nil{return "",err};return hex.EncodeToString(b),nil}
func Start(ctx context.Context,operation,traceID string,_ int)(context.Context,*Span){if ctx==nil{ctx=context.Background()};if traceID==""{if id,err:=newID(16);err==nil{traceID=id}else{traceID="00000000000000000000000000000000"}};spanID,idErr:=newID(8);if idErr!=nil{spanID="0000000000000000"};s:=&Span{operation:operation,traceID:traceID,spanID:spanID,start:time.Now().UTC()};return context.WithValue(ctx,traceKey{},traceID),s}
func TraceID(ctx context.Context)string{if ctx==nil{return ""};if id,ok:=ctx.Value(traceKey{}).(string);ok{return id};return ""}
func End(s *Span,err error){if s==nil{return};s.mu.Lock();if s.ended{s.mu.Unlock();return};s.ended=true;s.end=time.Now().UTC();if err!=nil{s.err=err.Error()};record:=TraceRecord{TraceID:s.traceID,SpanID:s.spanID,Operation:s.operation,StartedAt:s.start,EndedAt:s.end,Duration:s.end.Sub(s.start),Error:s.err};s.mu.Unlock();globalRecorder.mu.Lock();globalRecorder.records=append(globalRecorder.records,record);if len(globalRecorder.records)>globalRecorder.capacity{globalRecorder.records=globalRecorder.records[len(globalRecorder.records)-globalRecorder.capacity:]};globalRecorder.mu.Unlock()}
func SnapshotTraces()[]TraceRecord{globalRecorder.mu.RLock();defer globalRecorder.mu.RUnlock();out:=make([]TraceRecord,len(globalRecorder.records));copy(out,globalRecorder.records);return out}
func ResetTraces(){globalRecorder.mu.Lock();globalRecorder.records=nil;globalRecorder.mu.Unlock()}
func ValidateTraceID(id string)error{if len(id)!=32{return errors.New("trace id must contain 16 bytes encoded as 32 hex characters")};if _,err:=hex.DecodeString(id);err!=nil{return errors.New("trace id must be hexadecimal")};return nil}
