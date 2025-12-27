package protocol

// Intent represents a player input action as a bitmask
type Intent uint8

const (
	IntentNone   Intent = 0
	IntentLeft   Intent = 1 << iota
	IntentRight
	IntentJump
	IntentAttack
	IntentUse
)

// InputFrame contains player input for a single tick
type InputFrame struct {
	Tick    uint64
	Intents Intent
}

// EntityID uniquely identifies an entity
type EntityID uint64

// EntityState is the serialized state of an entity
type EntityState struct {
	ID         EntityID
	Components []byte // Serialized via ark-serde
}

// StateSnapshot contains game state for a tick
type StateSnapshot struct {
	Tick     uint64
	Full     bool     // True = complete state, False = delta
	Baseline uint64   // If delta, relative to this tick
	Entities []EntityState
	Removed  []EntityID // Entities removed since baseline
}

// Handshake is exchanged on connection
type Handshake struct {
	Version    int
	PlayerName string
}

// Message types for network protocol
type MsgType uint8

const (
	MsgHandshake MsgType = iota
	MsgInput
	MsgState
	MsgTick
	MsgPing
	MsgPong
	MsgDisconnect
)
