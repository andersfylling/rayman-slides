# protocol

Shared types for client-server communication. This package has no dependencies on other internal packages.

## Contents

- **version.go** - Protocol version constants for compatibility checks
- **messages.go** - Message types, intents, entity state

## Key Types

```go
// Player input as a bitmask
type Intent uint8
const (
    IntentLeft Intent = 1 << iota
    IntentRight
    IntentJump
    IntentAttack
    IntentUse
)

// Input for one tick
type InputFrame struct {
    Tick    uint64
    Intents Intent
}

// Game state snapshot
type StateSnapshot struct {
    Tick     uint64
    Full     bool
    Entities []EntityState
    Removed  []EntityID
}
```

## Version Compatibility

Client and server exchange versions on connect. Incompatible versions reject the connection.

```go
if !protocol.Compatible(localVersion, remoteVersion) {
    return ErrVersionMismatch
}
```
