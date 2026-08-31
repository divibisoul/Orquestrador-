package grpcapi

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
)

const (
	ServiceName = "nexus.v1.Orchestrator"
	CreateWorkflowMethod = "/nexus.v1.Orchestrator/CreateWorkflow"
	GetWorkflowStatusMethod = "/nexus.v1.Orchestrator/GetWorkflowStatus"
)

type WorkflowTask struct {
	TaskID string `json:"task_id"`
	Goal string `json:"goal"`
	Priority int32 `json:"priority,omitempty"`
	PrecisionRequired string `json:"precision_required,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type WorkflowRequest struct {
	WorkflowID string `json:"workflow_id"`
	Tasks []WorkflowTask `json:"tasks,omitempty"`
}

type WorkflowReply struct {
	WorkflowID string `json:"workflow_id"`
	Status string `json:"status"`
	Error string `json:"error,omitempty"`
}

type OrchestratorServer interface {
	CreateWorkflow(context.Context, *WorkflowRequest) (*WorkflowReply, error)
	GetWorkflowStatus(context.Context, *WorkflowRequest) (*WorkflowReply, error)
}

type Server struct { Engine *orchestrator.Engine }

func NewServer(engine *orchestrator.Engine) *Server {
	if engine == nil { engine = orchestrator.NewEngine(4) }
	return &Server{Engine: engine}
}

func (s *Server) CreateWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowReply, error) {
	if s == nil || s.Engine == nil { return nil, status.Error(codes.FailedPrecondition, "orchestrator engine unavailable") }
	if req == nil || strings.TrimSpace(req.WorkflowID) == "" { return nil, status.Error(codes.InvalidArgument, "workflow_id required") }
	steps := make([]orchestrator.Step, 0, len(req.Tasks))
	for _, task := range req.Tasks {
		id := strings.TrimSpace(task.TaskID)
		if id == "" { return nil, status.Error(codes.InvalidArgument, "task_id required") }
		goal := strings.TrimSpace(task.Goal)
		if goal == "" { goal = id }
		steps = append(steps, orchestrator.Step{ID: id, Run: func(context.Context) error { return nil }, Compensation: func(context.Context) error { return nil }})
		_ = goal
	}
	if len(steps) == 0 { return nil, status.Error(codes.InvalidArgument, "at least one task required") }
	if err := s.Engine.CreateWorkflow(req.WorkflowID, steps); err != nil { return nil, status.Error(codes.AlreadyExists, err.Error()) }
	return &WorkflowReply{WorkflowID: req.WorkflowID, Status: "pending"}, nil
}

func (s *Server) GetWorkflowStatus(_ context.Context, req *WorkflowRequest) (*WorkflowReply, error) {
	if s == nil || s.Engine == nil { return nil, status.Error(codes.FailedPrecondition, "orchestrator engine unavailable") }
	if req == nil || strings.TrimSpace(req.WorkflowID) == "" { return nil, status.Error(codes.InvalidArgument, "workflow_id required") }
	statusValue, err := s.Engine.GetWorkflowStatus(req.WorkflowID)
	if err != nil { return nil, status.Error(codes.NotFound, err.Error()) }
	return &WorkflowReply{WorkflowID: req.WorkflowID, Status: string(statusValue)}, nil
}

func Register(server *grpc.Server, impl OrchestratorServer) error {
	if server == nil { return errors.New("nil gRPC server") }
	if impl == nil { return errors.New("nil orchestrator service") }
	encoding.RegisterCodec(jsonCodec{})
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: ServiceName,
		HandlerType: (*OrchestratorServer)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "CreateWorkflow", Handler: createWorkflowHandler},
			{MethodName: "GetWorkflowStatus", Handler: getWorkflowStatusHandler},
		},
		Streams: nil,
		Metadata: "api/proto/orchestrator.proto",
	}, impl)
	return nil
}

func Serve(lis net.Listener, engine *orchestrator.Engine) (*grpc.Server, error) {
	if lis == nil { return nil, errors.New("nil listener") }
	server := grpc.NewServer()
	if err := Register(server, NewServer(engine)); err != nil { return nil, err }
	healthServer := health.NewServer()
	healthServer.SetServingStatus(ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	go func(){ _ = server.Serve(lis) }()
	return server, nil
}

func createWorkflowHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkflowRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(OrchestratorServer).CreateWorkflow(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: CreateWorkflowMethod}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(OrchestratorServer).CreateWorkflow(ctx, req.(*WorkflowRequest)) }
	return interceptor(ctx, in, info, handler)
}

func getWorkflowStatusHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkflowRequest)
	if err := dec(in); err != nil { return nil, err }
	if interceptor == nil { return srv.(OrchestratorServer).GetWorkflowStatus(ctx, in) }
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: GetWorkflowStatusMethod}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return srv.(OrchestratorServer).GetWorkflowStatus(ctx, req.(*WorkflowRequest)) }
	return interceptor(ctx, in, info, handler)
}

type jsonCodec struct{}
func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(v interface{}) ([]byte,error){ return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte,v interface{})error{return json.Unmarshal(data,v)}
