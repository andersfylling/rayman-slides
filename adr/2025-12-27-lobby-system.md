# Lobby System

**Status:** Accepted

## Context

Multiplayer game needs a way for players to find and join games. Must balance ease of use, infrastructure costs, and implementation complexity for MVP.

## Options Considered

### Direct Connect

Player manually enters server IP and port.

**Workflow:** Host starts server → shares IP:port → others type it in → connect

**Pros:**
- Zero infrastructure needed
- Simple to implement
- Works for LAN parties
- Full player control

**Cons:**
- Requires knowing IP (technical barrier)
- NAT/firewall issues for home hosting
- No discovery mechanism
- Poor UX for casual players

### Server Browser

Central service lists active servers. Players browse and pick.

**Workflow:** Server registers with master → master maintains list → client fetches list → pick server

**Pros:**
- Discoverable servers
- Can show ping, player count, map
- Familiar pattern (classic multiplayer)

**Cons:**
- Requires always-on master server
- Master server is single point of failure
- Spam/fake server risk
- Ongoing hosting cost

### Matchmaking

Central service automatically pairs players into games.

**Workflow:** Player queues → matchmaker finds others → spins up server → connects players

**Pros:**
- Best UX (just click "play")
- Can match by skill/region
- No manual server hunting

**Cons:**
- Most complex to implement
- Requires always-on matchmaker + server pool
- Highest infrastructure cost
- Overkill for small player base

### Room Codes

Generate short alphanumeric code. Share via chat/voice.

**Workflow:** Host creates room → gets code (e.g., XKCD-1234) → shares → others enter code → connect

**Pros:**
- Simple UX (just share code)
- Minimal infrastructure (code → IP lookup)
- Works well with Discord/friends
- NAT traversal solvable with relay

**Cons:**
- Requires lightweight lookup service
- Codes must expire/recycle
- No public discovery (intentional?)
- Still need relay for NAT traversal

### Platform Integration

Use Steam, Discord, Epic lobbies and invites.

**Workflow:** Integrate SDK → use platform's lobby/invite system → platform handles discovery

**Pros:**
- Polished UX (native invites, rich presence)
- NAT traversal handled
- Friends list integration
- No custom infrastructure

**Cons:**
- Platform lock-in
- SDK integration complexity
- May require store listing/approval
- Multiple platforms = multiple integrations
- Terminal game may not fit platform model

## Decision

**Tiered approach: Direct connect + Room codes, server browser later:**

### Tier 1: MVP
1. **Direct connect** - Always available, works for LAN and technical users
2. **Room codes** - Simple code lookup service for friends playing together

### Tier 2: Growth
3. **Server browser** - Add when player base justifies infrastructure

### Tier 3: Polish
4. **Platform integration** - Discord Rich Presence / invites (low-hanging fruit)

Room code implementation:
```
1. Host starts server (local or dedicated)
2. Server registers with lookup service: {code: "ABCD-1234", ip: "x.x.x.x", port: 7777}
3. Host shares code out-of-band (Discord, text, voice)
4. Client enters code → lookup service returns IP:port → client connects
5. Code expires after N hours or when server shuts down
```

Lookup service can be:
- Simple HTTP API (cheap to host)
- Cloudflare Workers / Vercel edge function (nearly free)
- Redis with TTL for code expiry

```
lobby/
  direct.go       # Direct IP:port connection
  roomcode.go     # Code generation and lookup client
  browser.go      # Server browser (future)

# Separate service
cmd/
  lookup/         # Room code lookup service
```

Room code format: `WORD-NNNN` or `XXXX-XXXX`
- Easy to read aloud
- 4 alphanumeric = 1.6M combinations (plenty)
- Hyphen prevents ambiguity

```go
type Room struct {
    Code      string    `json:"code"`
    Host      string    `json:"host"`      // IP:port
    Name      string    `json:"name"`      // Optional display name
    Players   int       `json:"players"`
    MaxPlayers int      `json:"max_players"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

## Consequences

- **Low barrier MVP** - Direct connect works day one, no infrastructure
- **Friend-friendly** - Room codes easy to share in Discord/chat
- **Minimal cost** - Lookup service is tiny, cheap/free hosting options
- **NAT not solved** - Players behind strict NAT need port forwarding (revisit with relay later)
- **No matchmaking** - Players must coordinate externally (fine for niche game)
- **Extensible** - Server browser and platform integration can layer on later
