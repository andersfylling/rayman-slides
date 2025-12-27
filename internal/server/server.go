// Package server implements the authoritative game server.
// Can be embedded in the client for local play or run standalone.
package server

import (
	"sync"
)

// Config holds server configuration
type Config struct {
	Port       int
	MaxPlayers int
	TickRate   int // Ticks per second
	MapPath    string
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Port:       7777,
		MaxPlayers: 4,
		TickRate:   60,
		MapPath:    "",
	}
}

// Server is the authoritative game server
type Server struct {
	config   Config
	tick     uint64
	running  bool
	mu       sync.RWMutex
	// TODO: ECS world
	// TODO: Sessions map
	// TODO: Network listener
}

// New creates a new server with the given config
func New(cfg Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Start begins the server tick loop
func (s *Server) Start() error {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	// TODO: Load map
	// TODO: Initialize ECS
	// TODO: Start network listener
	// TODO: Run tick loop

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

// Tick returns the current tick number
func (s *Server) Tick() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tick
}
