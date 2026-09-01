package grpcapi

import (
	"context"

	"google.golang.org/grpc"
)

type Client struct{ conn *grpc.ClientConn }

func Dial(ctx context.Context, target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil { return nil, err }
	return &Client{conn: conn}, nil
}

func NewClientConn(conn *grpc.ClientConn) *Client { if conn == nil { return nil }; return &Client{conn: conn} }
func (c *Client) Close() error { if c == nil || c.conn == nil { return nil }; return c.conn.Close() }
func (c *Client) CreateWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowReply, error) {
	out := new(WorkflowReply)
	if err := c.conn.Invoke(ctx, CreateWorkflowMethod, req, out, grpc.ForceCodec(jsonCodec{})); err != nil { return nil, err }
	return out, nil
}
func (c *Client) GetWorkflowStatus(ctx context.Context, req *WorkflowRequest) (*WorkflowReply, error) {
	out := new(WorkflowReply)
	if err := c.conn.Invoke(ctx, GetWorkflowStatusMethod, req, out, grpc.ForceCodec(jsonCodec{})); err != nil { return nil, err }
	return out, nil
}
