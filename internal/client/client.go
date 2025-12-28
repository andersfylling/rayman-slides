// Package client implements the game client.
// The client connects input, server, and renderer.
package client

import (
	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/input"
	"github.com/andersfylling/rayman-slides/internal/protocol"
	"github.com/andersfylling/rayman-slides/internal/server"
)

// Client connects input system to server and provides state for rendering.
// In single-player, the server runs locally (embedded).
// In multiplayer, inputs are also sent to an external server.
type Client struct {
	playerID  int
	sessionID int

	// Input state
	keyState *input.KeyState

	// Internal server (always runs locally for prediction)
	server *server.Server

	// External server connection (nil for single-player)
	// TODO: externalConn *network.Connection

	// State for multiplayer sync
	lastSentTick uint64
}

// New creates a new client.
func New(playerID int) *Client {
	return &Client{
		playerID:  playerID,
		sessionID: 1, // Single-player uses session 1
		keyState:  input.NewKeyState(),
	}
}

// SetServer sets the internal server.
func (c *Client) SetServer(s *server.Server) {
	c.server = s
	// Register ourselves as a session
	c.server.AddSession(c.sessionID, c.playerID, "Player")
}

// ProcessInput handles input events and sends them to the server.
func (c *Client) ProcessInput(events []input.KeyEvent) {
	// Update local key state
	for _, ev := range events {
		switch ev.Type {
		case input.KeyDown:
			c.keyState.SetPressed(ev.Key, true)
		case input.KeyUp:
			c.keyState.SetPressed(ev.Key, false)
		}
	}

	// Convert to intents and send to server
	intents := c.keyState.ToIntents()
	tick := c.server.Tick()

	frame := protocol.InputFrame{
		Tick:    tick + 1, // Input is for the next tick
		Intents: intents,
	}

	// Send to internal server (local prediction)
	c.server.QueueInput(c.sessionID, frame)

	// TODO: Also send to external server for multiplayer
	// if c.externalConn != nil {
	//     c.externalConn.Send(frame)
	// }
}

// World returns the current game world state (for rendering).
func (c *Client) World() *game.World {
	return c.server.World()
}

// ShouldQuit checks if quit was requested.
func (c *Client) ShouldQuit() bool {
	return c.keyState.IsPressed(input.KeyQuit)
}
