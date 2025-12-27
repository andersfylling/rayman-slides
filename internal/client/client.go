// Package client implements the game client.
// Handles rendering, input capture, and network communication.
package client

import (
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Config holds client configuration
type Config struct {
	ServerAddr string // Empty for local/embedded server
	PlayerName string
	RenderMode RenderMode
}

// RenderMode specifies the terminal rendering approach
type RenderMode int

const (
	RenderAuto      RenderMode = iota // Auto-detect best mode
	RenderASCII                       // Plain ASCII
	RenderHalfBlock                   // Half-block with color
	RenderBraille                     // Braille patterns
)

// Client is the game client
type Client struct {
	config      Config
	connected   bool
	serverTick  uint64
	inputBuffer []protocol.InputFrame
	// TODO: Renderer
	// TODO: Input handler
	// TODO: Network connection
	// TODO: State interpolation buffer
}

// New creates a new client with the given config
func New(cfg Config) *Client {
	return &Client{
		config:      cfg,
		inputBuffer: make([]protocol.InputFrame, 0, 8),
	}
}

// Connect connects to a remote server or starts embedded server
func (c *Client) Connect() error {
	// TODO: If ServerAddr empty, start embedded server
	// TODO: Otherwise, dial remote server
	// TODO: Perform handshake
	return nil
}

// Run starts the client main loop
func (c *Client) Run() error {
	// TODO: Initialize terminal
	// TODO: Start input capture
	// TODO: Start render loop
	// TODO: Process network messages
	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() {
	c.connected = false
	// TODO: Send disconnect message
	// TODO: Close network connection
}
