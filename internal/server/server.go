// Package server implements the authoritative game server.
// Can be embedded in the client for local play or run standalone.
package server

import (
	"sync"
	"time"

	"github.com/andersfylling/rayman-slides/internal/game"
	"github.com/andersfylling/rayman-slides/internal/protocol"
)

// Config holds server configuration
type Config struct {
	Port       int
	MaxPlayers int
	TickRate   int           // Game ticks per second
	SyncRate   int           // State broadcasts per second (can be lower than tick rate)
	MapPath    string
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Port:       7777,
		MaxPlayers: 4,
		TickRate:   60,
		SyncRate:   20, // Broadcast state 20 times per second
		MapPath:    "",
	}
}

// Session represents a connected client
type Session struct {
	ID          int
	PlayerID    int
	Name        string
	InputQueue  []protocol.InputFrame // Pending inputs to process
	LastAckTick uint64                // Last tick acknowledged by client
	mu          sync.Mutex
}

// QueueInput adds an input frame to the session's queue
func (s *Session) QueueInput(frame protocol.InputFrame) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InputQueue = append(s.InputQueue, frame)
}

// DrainInputs returns and clears all pending inputs up to the given tick
func (s *Session) DrainInputs(upToTick uint64) []protocol.InputFrame {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []protocol.InputFrame
	remaining := s.InputQueue[:0]

	for _, input := range s.InputQueue {
		if input.Tick <= upToTick {
			result = append(result, input)
		} else {
			remaining = append(remaining, input)
		}
	}

	s.InputQueue = remaining
	return result
}

// Server is the authoritative game server
type Server struct {
	config   Config
	tick     uint64
	running  bool
	mu       sync.RWMutex

	world    *game.World
	sessions map[int]*Session // sessionID -> session

	// Channels
	quitCh   chan struct{}
	doneCh   chan struct{}

	// Callbacks for embedded mode (when server runs in same process as client)
	onStateUpdate func(state game.WorldState)
}

// New creates a new server with the given config
func New(cfg Config) *Server {
	return &Server{
		config:   cfg,
		sessions: make(map[int]*Session),
		quitCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// SetWorld sets the game world (for embedded mode where client creates the world)
func (s *Server) SetWorld(w *game.World) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.world = w
}

// World returns the server's game world
func (s *Server) World() *game.World {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.world
}

// SetStateUpdateCallback sets a callback for state updates (embedded mode)
func (s *Server) SetStateUpdateCallback(cb func(state game.WorldState)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onStateUpdate = cb
}

// AddSession adds a new session for a connected client
func (s *Server) AddSession(sessionID int, playerID int, name string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &Session{
		ID:         sessionID,
		PlayerID:   playerID,
		Name:       name,
		InputQueue: make([]protocol.InputFrame, 0, 16),
	}
	s.sessions[sessionID] = session
	return session
}

// RemoveSession removes a session
func (s *Server) RemoveSession(sessionID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// QueueInput adds an input to a session's queue
func (s *Server) QueueInput(sessionID int, frame protocol.InputFrame) {
	s.mu.RLock()
	session, ok := s.sessions[sessionID]
	s.mu.RUnlock()

	if ok {
		session.QueueInput(frame)
	}
}

// Start begins the server tick loop
func (s *Server) Start() error {
	s.mu.Lock()
	if s.world == nil {
		s.world = game.NewWorld()
	}
	s.running = true
	s.mu.Unlock()

	go s.runTickLoop()

	return nil
}

// StartBlocking runs the tick loop on the current goroutine
func (s *Server) StartBlocking() error {
	s.mu.Lock()
	if s.world == nil {
		s.world = game.NewWorld()
	}
	s.running = true
	s.mu.Unlock()

	s.runTickLoop()
	return nil
}

func (s *Server) runTickLoop() {
	defer close(s.doneCh)

	tickDuration := time.Second / time.Duration(s.config.TickRate)
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	// Sync rate for state broadcasts
	syncInterval := s.config.TickRate / s.config.SyncRate
	if syncInterval < 1 {
		syncInterval = 1
	}
	ticksSinceSync := 0

	for {
		select {
		case <-s.quitCh:
			return
		case <-ticker.C:
			s.processTick()

			// Broadcast state at sync rate
			ticksSinceSync++
			if ticksSinceSync >= syncInterval {
				ticksSinceSync = 0
				s.broadcastState()
			}
		}
	}
}

func (s *Server) processTick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Collect and apply inputs from all sessions
	for _, session := range s.sessions {
		inputs := session.DrainInputs(s.tick + 1)
		for _, input := range inputs {
			// Apply input to player entity
			s.world.SetPlayerIntent(session.PlayerID, input.Intents)
		}
	}

	// Run game simulation
	s.world.Update()
	s.tick = s.world.Tick
}

func (s *Server) broadcastState() {
	s.mu.RLock()
	state := s.world.Snapshot()
	callback := s.onStateUpdate
	s.mu.RUnlock()

	// For embedded mode, call the callback directly
	if callback != nil {
		callback(state)
	}

	// TODO: For network mode, serialize and send to all sessions
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.quitCh)
	<-s.doneCh
}

// Tick returns the current tick number
func (s *Server) Tick() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tick
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
