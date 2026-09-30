package mesh

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/divibisoul/Orquestrador-/protocol"
)

const ClareiraCapability = "clareira.ingest"

type ClareiraEvent struct {
	Phase string
	Operation string
	DeviceID string
	Backend string
	InputSize int
	OutputSize int
	CorrelationID string
	Error string
}

type PeerCaller interface {
	CallWithCorrelation(context.Context, string, string, map[string]any, string) (map[string]any, error)
}

type ClareiraReporter struct{ peers PeerCaller }

func NewClareiraReporter(peers PeerCaller)(*ClareiraReporter,error){
	if peers==nil{return nil,errors.New("Clareira reporter requires Mesh peer client")}
	return &ClareiraReporter{peers:peers},nil
}

func (r *ClareiraReporter) Report(ctx context.Context,event ClareiraEvent) error{
	if ctx==nil{return errors.New("context is nil")}
	cid:=strings.TrimSpace(event.CorrelationID);if cid==""{cid=protocol.NewTraceID()}
	raw,err:=json.Marshal(map[string]any{
		"phase":event.Phase,"operation":event.Operation,"device_id":event.DeviceID,"backend":event.Backend,
		"input_size":event.InputSize,"output_size":event.OutputSize,"correlation_id":cid,"error":event.Error,
	})
	if err!=nil{return err}
	packet:=map[string]any{
		"id":protocol.NewTraceID(),"data":string(raw),"informationalValue":float64(event.OutputSize),
		"criticality":criticalityForPhase(event.Phase),"packetType":"StateReport","sourceId":"N07","timestamp":time.Now().UnixMilli(),
		"correlationId":cid,
	}
	_,err=r.peers.CallWithCorrelation(ctx,"N01",ClareiraCapability,map[string]any{"packet":packet},cid)
	return err
}

func criticalityForPhase(phase string)float64{
	switch strings.ToLower(strings.TrimSpace(phase)){case "failed":return 1;case "started":return .6;default:return .4}
}
