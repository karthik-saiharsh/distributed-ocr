package rpc

import (
	"fmt"
	"net/rpc"
)

// Client wraps a standard net/rpc Client for convenience.
type Client struct {
	addr string
	conn *rpc.Client
}

// NewClient creates a new RPC client connected to the given address.
func NewClient(addr string) (*Client, error) {
	conn, err := rpc.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC at %s: %w", addr, err)
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// AssignTask sends a TaskRequest to a Worker Node for processing.
func (c *Client) AssignTask(req TaskRequest) (*TaskResponse, error) {
	var resp TaskResponse
	err := c.conn.Call("WorkerRPC.AssignTask", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("RPC AssignTask failed: %w", err)
	}
	return &resp, nil
}

// StealTask sends a StealRequest to a Worker Node to steal tasks.
func (c *Client) StealTask(req StealRequest) (*StealResponse, error) {
	var resp StealResponse
	err := c.conn.Call("WorkerRPC.StealTask", req, &resp)
	if err != nil {
		return nil, fmt.Errorf("RPC StealTask failed: %w", err)
	}
	return &resp, nil
}

// SubmitStolenResult returns a completed task response to the original victim node.
func (c *Client) SubmitStolenResult(req TaskResponse) (bool, error) {
	var ack bool
	err := c.conn.Call("WorkerRPC.SubmitStolenResult", req, &ack)
	if err != nil {
		return false, fmt.Errorf("RPC SubmitStolenResult failed: %w", err)
	}
	return ack, nil
}

// SendHeartbeat sends an application-layer keepalive to the Master.
func (c *Client) SendHeartbeat(req HeartbeatRequest) (bool, error) {
	var resp bool
	err := c.conn.Call("MasterRPC.Heartbeat", req, &resp)
	if err != nil {
		return false, fmt.Errorf("RPC SendHeartbeat failed: %w", err)
	}
	return resp, nil
}
