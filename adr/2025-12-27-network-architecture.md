# Network Architecture

**Status:** Accepted

## Context

Multiplayer platformer needs networking. Already decided on tick-based server authority (see game-loop ADR). Must choose network topology and transport protocol.

## Options Considered: Topology

### Client-Server (Authoritative)

Central server runs game logic. Clients send inputs, receive state.

**Workflow:** Client input → server → simulate → broadcast state → clients render

**Pros:**
- Single source of truth (no desync)
- Cheat resistant (server validates everything)
- Simple client logic (just render)
- Easy to host dedicated servers

**Cons:**
- Server is single point of failure
- Latency to server affects all players
- Server costs for hosting

### Peer-to-Peer

No central server. Peers connect directly and sync state.

**Workflow:** Each peer simulates → exchange state → resolve conflicts

**Pros:**
- No server costs
- Lower latency between nearby peers
- Works without infrastructure

**Cons:**
- NAT traversal is painful
- No authority = desync risk
- Cheat-prone (each client has full state)
- Complex conflict resolution
- Host migration if "host" leaves

### Lockstep

All clients simulate identically. Only exchange inputs.

**Workflow:** Collect all player inputs → everyone simulates same tick → repeat

**Pros:**
- Minimal bandwidth (only inputs sent)
- Guaranteed sync if deterministic
- Good for strategy games (RTS)

**Cons:**
- Slowest player delays everyone
- Requires perfect determinism (floating point issues)
- High latency feels terrible for action games
- One dropped packet stalls all players

### Relay Server

Server forwards packets but doesn't run game logic.

**Workflow:** Client → relay → other clients (relay doesn't process)

**Pros:**
- Simple server (just routing)
- NAT traversal solved
- Lower server CPU

**Cons:**
- Still need authority somewhere (one client = host?)
- Host advantage / cheat risk
- Combines downsides of P2P and client-server

## Options Considered: Protocol

### TCP

Reliable, ordered byte stream.

**Pros:**
- Guaranteed delivery and order
- Simple to program
- Works through most firewalls

**Cons:**
- Head-of-line blocking (one lost packet stalls all)
- Higher latency for real-time
- No partial delivery

### UDP

Unreliable datagrams.

**Pros:**
- Low latency (no retransmit wait)
- Can drop stale packets intentionally
- Full control over reliability

**Cons:**
- Must implement own reliability layer
- Packet loss handling complexity
- Some firewalls/NATs unfriendly

### WebSocket

TCP-based, framed messages, HTTP upgrade.

**Pros:**
- Works in browsers
- Easy firewall traversal
- Good library support
- Framed (no manual message boundary handling)

**Cons:**
- Still TCP underneath (head-of-line blocking)
- WebSocket overhead per frame
- Overkill if not targeting browsers

### QUIC

UDP-based, multiplexed streams, TLS built-in.

**Pros:**
- Multiple streams without head-of-line blocking
- Fast connection establishment (0-RTT)
- Built-in encryption
- Modern protocol, designed for games/real-time

**Cons:**
- Newer, less library maturity
- More complex than raw UDP
- Some networks block/throttle UDP

## Decision

**Client-server authoritative over TCP, with upgrade path to QUIC:**

1. **Client-server** - Matches our tick-based server authority model
2. **TCP initially** - Simple, reliable, good enough for terminal game tick rates
3. **QUIC later** - Upgrade path if TCP latency becomes problematic
4. **WebSocket optional** - Add if browser client ever wanted

Rationale: Terminal game at 30-60 ticks/sec is forgiving. TCP's reliability saves implementation time. QUIC can be swapped in later since message format is independent of transport.

```
network/
  protocol/
    messages.go   # Message types (InputFrame, StateSnapshot, etc.)
    codec.go      # Serialization (msgpack or protobuf)
  transport/
    tcp.go        # TCP implementation
    quic.go       # QUIC implementation (future)
  server/
    server.go     # Accept connections, manage sessions
    session.go    # Per-client state
  client/
    client.go     # Connect, send input, receive state
```

Message types:
```go
// Client → Server
MsgInput        // InputFrame for a tick

// Server → Client
MsgState        // Full or delta state snapshot
MsgTick         // Tick acknowledgment / timing

// Both directions
MsgHandshake    // Version check, player info
MsgPing         // Latency measurement
MsgDisconnect   // Graceful disconnect
```

## Consequences

- **Server authority** - Aligns with existing game loop decision
- **TCP simplicity** - No custom reliability code, faster to implement
- **Latency acceptable** - Terminal tick rate is forgiving
- **Transport abstraction** - Can swap TCP → QUIC without changing game code
- **No browser support yet** - WebSocket deferred, add if needed
- **Server hosting** - Need to run/pay for servers (or players self-host)
