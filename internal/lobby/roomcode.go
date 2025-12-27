// Package lobby handles game discovery and joining.
package lobby

import (
	"fmt"
	"math/rand"
	"time"
)

// Room represents a game room
type Room struct {
	Code       string    `json:"code"`
	Host       string    `json:"host"` // IP:port
	Name       string    `json:"name"`
	Players    int       `json:"players"`
	MaxPlayers int       `json:"max_players"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// CodeGenerator generates room codes
type CodeGenerator struct {
	rng *rand.Rand
}

// NewCodeGenerator creates a code generator
func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Generate creates a new room code in format XXXX-XXXX
func (g *CodeGenerator) Generate() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No I, O, 0, 1 (ambiguous)
	code := make([]byte, 9)
	for i := 0; i < 4; i++ {
		code[i] = charset[g.rng.Intn(len(charset))]
	}
	code[4] = '-'
	for i := 5; i < 9; i++ {
		code[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(code)
}

// RoomStore stores active rooms (in-memory implementation)
type RoomStore struct {
	rooms map[string]*Room
	ttl   time.Duration
}

// NewRoomStore creates a room store
func NewRoomStore(ttl time.Duration) *RoomStore {
	return &RoomStore{
		rooms: make(map[string]*Room),
		ttl:   ttl,
	}
}

// Create creates a new room and returns the code
func (s *RoomStore) Create(host, name string, maxPlayers int) (*Room, error) {
	gen := NewCodeGenerator()
	code := gen.Generate()

	// Ensure unique (retry if collision)
	for i := 0; i < 10; i++ {
		if _, exists := s.rooms[code]; !exists {
			break
		}
		code = gen.Generate()
	}

	room := &Room{
		Code:       code,
		Host:       host,
		Name:       name,
		Players:    1,
		MaxPlayers: maxPlayers,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(s.ttl),
	}
	s.rooms[code] = room
	return room, nil
}

// Lookup finds a room by code
func (s *RoomStore) Lookup(code string) (*Room, error) {
	room, exists := s.rooms[code]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", code)
	}
	if time.Now().After(room.ExpiresAt) {
		delete(s.rooms, code)
		return nil, fmt.Errorf("room expired: %s", code)
	}
	return room, nil
}

// Delete removes a room
func (s *RoomStore) Delete(code string) {
	delete(s.rooms, code)
}

// Cleanup removes expired rooms
func (s *RoomStore) Cleanup() {
	now := time.Now()
	for code, room := range s.rooms {
		if now.After(room.ExpiresAt) {
			delete(s.rooms, code)
		}
	}
}
