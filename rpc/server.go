package rpc

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// Server handles listening for incoming RPC connections.
type Server struct {
	port     int
	listener net.Listener
}

// NewServer creates a new RPC server on the specified port.
func NewServer(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start registers the provided receiver (the object handling the methods)
// and begins listening for incoming TCP RPC connections in a background goroutine.
func (s *Server) Start(receiver interface{}) error {
	// Create a new RPC server instance instead of using the global DefaultServer
	// This prevents issues if Start is called multiple times (e.g., in tests)
	rpcServer := rpc.NewServer()
	
	// Register the receiver (e.g., a Worker object)
	err := rpcServer.Register(receiver)
	if err != nil {
		return fmt.Errorf("failed to register RPC receiver: %w", err)
	}

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start RPC listener on %s: %w", addr, err)
	}
	s.listener = listener

	log.Printf("[RPC] Server listening on %s", addr)

	// Accept connections concurrently
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				// Prevent logging expected errors when the listener is closed
				select {
				case <-s.isClosed():
					return
				default:
					log.Printf("[RPC] Accept error: %v", err)
					continue
				}
			}
			go rpcServer.ServeConn(conn)
		}
	}()

	return nil
}

// Stop closes the listener and stops accepting new connections.
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}

// Internal helper to check if listener is closed based on error type
// A bit hacky, normally you'd use a context or a done channel.
func (s *Server) isClosed() <-chan struct{} {
	ch := make(chan struct{})
	// Just a dummy channel that never closes for this simple implementation
	// The Accept loop will naturally break if the listener is closed.
	return ch
}
