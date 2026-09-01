package mesh

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/orchestrator"
	"github.com/divibisoul/Orquestrador-/protocol"
)

type HTTPGateway struct { Engine *orchestrator.Engine; Secret string; AllowUnauthenticatedLocal bool }

func NewHTTPGateway(engine *orchestrator.Engine) *HTTPGateway { return &HTTPGateway{Engine: engine, Secret: strings.TrimSpace(os.Getenv("SOUL_MESH_HMAC_SECRET")), AllowUnauthenticatedLocal: strings.EqualFold(strings.TrimSpace(os.Getenv("N07_MESH_ALLOW_UNAUTH_LOCAL")), "true")} }
func (g *HTTPGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) { g.Handler(w, r) }
func (g *HTTPGateway) authenticated(envelope protocol.MeshEnvelope) error { if g.Secret != "" { return protocol.VerifyHMAC(envelope, g.Secret, time.Now()) }; if g.AllowUnauthenticatedLocal { return nil }; return errors.New("Mesh HMAC secret is not configured") }
func (g *HTTPGateway) respond(w http.ResponseWriter, status int, in protocol.MeshEnvelope, typ string, payload map[string]any) { resp:=protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,MessageID:protocol.NewTraceID(),Source:"N07",Target:in.Source,Timestamp:time.Now().UnixMilli(),Nonce:protocol.NewTraceID(),CorrelationID:in.CorrelationID,Type:typ,Payload:payload}; if g.Secret!="" { if err:=protocol.SignHMAC(&resp,g.Secret); err!=nil { status=http.StatusInternalServerError; resp.Payload=map[string]any{"error":"response signing failed"} } }; writeMeshJSON(w,status,resp) }
func (g *HTTPGateway) Handler(w http.ResponseWriter,r *http.Request) { if r.Method!=http.MethodPost { writeMeshJSON(w,http.StatusMethodNotAllowed,protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,Type:"ERROR",Payload:map[string]any{"error":"POST required"}}); return }; if g.Engine==nil { writeMeshJSON(w,http.StatusServiceUnavailable,protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,Type:"ERROR",Payload:map[string]any{"error":"N07 engine unavailable"}}); return }; var envelope protocol.MeshEnvelope; if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)).Decode(&envelope); err!=nil { writeMeshJSON(w,http.StatusBadRequest,protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,Type:"ERROR",Payload:map[string]any{"error":err.Error()}}); return }; if err:=envelope.Validate(); err!=nil { writeMeshJSON(w,http.StatusBadRequest,gatewayError(envelope,err.Error())); return }; if err:=g.authenticated(envelope); err!=nil { g.respond(w,http.StatusUnauthorized,envelope,"ERROR",map[string]any{"error":"mesh authentication failed"}); return }; if envelope.Target!="N07"&&envelope.Target!="BROADCAST" { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"target is not N07"}); return }; capability:=envelope.Capability(); if envelope.Type=="PING"||capability=="mesh.ping" { g.respond(w,http.StatusOK,envelope,"TASK_RESULT",map[string]any{"ok":true,"nucleus":"N07"}); return }; if capability=="mesh.describe" { g.respond(w,http.StatusOK,envelope,"TASK_RESULT",map[string]any{"nucleus":"N07","operations":g.Engine.Operations(),"transports":[]string{"LOOPBACK_HTTP","HTTP"}}); return }; if capability=="" { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":"payload.capability is required"}); return }; values,err:=payloadValues(envelope.NestedPayload()); if err!=nil { g.respond(w,http.StatusBadRequest,envelope,"ERROR",map[string]any{"error":err.Error()}); return }; result,err:=g.Engine.ExecuteWithTrace(r.Context(),envelope.CorrelationID,envelope.Source,capability,values,envelope.NestedMetadata()); status:=http.StatusOK; if err!=nil { status=http.StatusBadRequest }; payload:=map[string]any{"values":result.Payload,"status":result.Status}; if result.Error!="" { payload["error"]=result.Error }; g.respond(w,status,envelope,"TASK_RESULT",payload) }
func payloadValues(payload map[string]any)([]float64,error){if payload==nil{return nil,errors.New("payload.values is required")};v,ok:=payload["values"];if !ok{return nil,errors.New("payload.values is required")};encoded,err:=json.Marshal(v);if err!=nil{return nil,err};var values []float64;if err:=json.Unmarshal(encoded,&values);err!=nil{return nil,errors.New("payload.values must be an array of numbers")};return values,nil}
func gatewayError(in protocol.MeshEnvelope,message string)protocol.MeshEnvelope{return protocol.MeshEnvelope{Version:protocol.SoulMeshVersion,ContractVersion:protocol.SoulMeshContractVersion,MessageID:protocol.NewTraceID(),Source:"N07",Target:in.Source,Timestamp:time.Now().UnixMilli(),Nonce:protocol.NewTraceID(),CorrelationID:in.CorrelationID,Type:"ERROR",Payload:map[string]any{"error":message}}}
func writeMeshJSON(w http.ResponseWriter,status int,envelope protocol.MeshEnvelope){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_=json.NewEncoder(w).Encode(envelope)}
func sorted(values []string)[]string{out:=append([]string(nil),values...);sort.Strings(out);return out}
