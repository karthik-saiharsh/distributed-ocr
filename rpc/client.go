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
