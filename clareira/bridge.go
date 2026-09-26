package clareira

import (
  "errors"
  "sort"
  "sync"
  "time"

  "github.com/divibisoul/Orquestrador-/shared"
)

const ContractVersion = "1.0.0"

type Metrics struct {
  CapturedAtMs int64 `json:"capturedAtMs"`
  Nodes struct { Total int `json:"total"`; Active int `json:"active"`; Errored int `json:"errored"` } `json:"nodes"`
  Channels struct { Total int `json:"total"`; Open int `json:"open"` } `json:"channels"`
  Packets struct { Ingested int64 `json:"ingested"`; Processed int64 `json:"processed"`; Dropped int64 `json:"dropped"`; Errored int64 `json:"errored"`; InFlight int64 `json:"inFlight"` } `json:"packets"`
  LatencyMs struct { Last float64 `json:"last"`; P50 float64 `json:"p50"`; P95 float64 `json:"p95"`; Max float64 `json:"max"` } `json:"latencyMs"`
  UptimeMs int64 `json:"uptimeMs"`
}

type Bridge struct { mu sync.Mutex; started int64; ingested, processed, dropped, errored, inFlight int64; latencies []float64 }
var Default = &Bridge{started:time.Now().UnixMilli()}

func (b *Bridge) Ingest(p shared.ClareiraPacket) (map[string]interface{}, error) {
  if p.ID=="" || p.Data=="" || p.SourceID=="" || p.CorrelationID=="" || p.Timestamp<=0 { return nil, errors.New("INVALID_CLAREIRA_PACKET") }
  if p.Criticality<0 || p.Criticality>1 { return nil, errors.New("INVALID_CLAREIRA_CRITICALITY") }
  switch p.PacketType { case "Data","StateReport","DecisionRequest","DecisionResponse","Control","Heartbeat": default: return nil, errors.New("INVALID_CLAREIRA_PACKET_TYPE") }
  b.mu.Lock(); b.ingested++; b.inFlight++; b.mu.Unlock()
  return map[string]interface{}{"accepted":true,"contractVersion":ContractVersion,"correlationId":p.CorrelationID,"sourceId":p.SourceID},nil
}

func (b *Bridge) Complete(started time.Time) {
  b.mu.Lock(); defer b.mu.Unlock(); b.processed++; if b.inFlight>0 { b.inFlight-- }; latency:=float64(time.Since(started).Microseconds())/1000; b.latencies=append(b.latencies,latency); if len(b.latencies)>128 { b.latencies=b.latencies[len(b.latencies)-128:] }
}

func (b *Bridge) Metrics() Metrics {
  b.mu.Lock(); defer b.mu.Unlock(); m:=Metrics{CapturedAtMs:time.Now().UnixMilli(),UptimeMs:time.Now().UnixMilli()-b.started}; m.Nodes.Total=1; m.Nodes.Active=1; m.Channels.Total=1; m.Channels.Open=1;
  if len(b.latencies)>0 { a:=append([]float64(nil),b.latencies...); sort.Float64s(a); pick:=func(q float64)float64{return a[int(float64(len(a)-1)*q)]}; m.LatencyMs.Last=a[len(a)-1]; m.LatencyMs.P50=pick(.5); m.LatencyMs.P95=pick(.95); m.LatencyMs.Max=a[len(a)-1] }
  m.Packets.Ingested=b.ingested; m.Packets.Processed=b.processed; m.Packets.Dropped=b.dropped; m.Packets.Errored=b.errored; m.Packets.InFlight=b.inFlight; return m
}
