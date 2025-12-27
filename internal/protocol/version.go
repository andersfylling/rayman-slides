// Package protocol defines shared types and version constants
// for client-server communication.
package protocol

// Version constants for compatibility checking
const (
	ProtocolVersion = 1
	MinVersion      = 1
)

// Compatible checks if two versions can communicate
func Compatible(local, remote int) bool {
	return remote >= MinVersion && local >= MinVersion
}
