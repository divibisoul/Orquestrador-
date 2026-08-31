package grpcapi

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/divibisoul/Orquestrador-/core/orchestrator"
)

func TestGRPCWorkflowRoundTrip(t *testing.T) {
	lis := bufconn.Listen(1 << 20)
	s := grpc.NewServer()
	if err := Register(s, NewServer(orchestrator.NewEngine(2))); err != nil { t.Fatal(err) }
	go s.Serve(lis)
	defer s.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil { t.Fatal(err) }
	defer conn.Close()

	client := NewClientConn(conn)
	reply, err := client.CreateWorkflow(ctx, &WorkflowRequest{WorkflowID: "grpc-e2e", Tasks: []WorkflowTask{{TaskID: "step-1", Goal: "execute"}}})
	if err != nil { t.Fatal(err) }
	if reply.WorkflowID != "grpc-e2e" || reply.Status != "pending" { t.Fatalf("reply=%+v", reply) }

	status, err := client.GetWorkflowStatus(ctx, &WorkflowRequest{WorkflowID: "grpc-e2e"})
	if err != nil { t.Fatal(err) }
	if status.Status != "pending" { t.Fatalf("status=%+v", status) }
}

func TestGRPCValidation(t *testing.T) {
	s := NewServer(orchestrator.NewEngine(1))
	if _, err := s.CreateWorkflow(context.Background(), nil); err == nil { t.Fatal("expected validation error") }
}
