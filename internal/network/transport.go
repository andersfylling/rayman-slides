// Package network implements client-server communication.
package network

import (
	"net"
)

// Transport abstracts the network connection
type Transport interface {
	// Connect establishes a connection to the server
	Connect(addr string) error

	// Accept waits for incoming connections (server only)
	Accept() (Connection, error)

	// Close closes the transport
	Close() error
}

// Connection represents a single client-server connection
type Connection interface {
	// Send sends a message
	Send(data []byte) error

	// Recv receives a message (blocking)
	Recv() ([]byte, error)

	// Close closes the connection
	Close() error

	// RemoteAddr returns the remote address
	RemoteAddr() net.Addr
}

// TCPTransport implements Transport over TCP
type TCPTransport struct {
	listener net.Listener
	conn     net.Conn
}

// NewTCPTransport creates a TCP transport
func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

// Listen starts listening on the given address (server)
func (t *TCPTransport) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	t.listener = ln
	return nil
}

// Connect connects to a server (client)
func (t *TCPTransport) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

// Accept accepts a new connection (server)
func (t *TCPTransport) Accept() (Connection, error) {
	conn, err := t.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &TCPConnection{conn: conn}, nil
}

// Close closes the transport
func (t *TCPTransport) Close() error {
	if t.listener != nil {
		return t.listener.Close()
	}
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// TCPConnection wraps a TCP connection
type TCPConnection struct {
	conn net.Conn
}

func (c *TCPConnection) Send(data []byte) error {
	// TODO: Length prefix for framing
	_, err := c.conn.Write(data)
	return err
}

func (c *TCPConnection) Recv() ([]byte, error) {
	// TODO: Read length prefix, then payload
	buf := make([]byte, 4096)
	n, err := c.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (c *TCPConnection) Close() error {
	return c.conn.Close()
}

func (c *TCPConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
