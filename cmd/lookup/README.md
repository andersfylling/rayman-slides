# lookup

Room code lookup service. A lightweight HTTP API that maps short room codes to server addresses.

## Why?

Sharing `192.168.1.100:7777` with friends is awkward. Room codes like `ABCD-1234` are easier to share over voice chat or text.

## Usage

```bash
# Start lookup service
./lookup --port 8080

# With custom TTL for rooms
./lookup --port 8080 --ttl 4h
```

## API

### Create Room

```bash
POST /rooms
Content-Type: application/json

{
  "host": "192.168.1.100:7777",
  "name": "My Game",
  "max_players": 4
}

# Response
{
  "code": "ABCD-1234",
  "expires_at": "2025-12-27T20:00:00Z"
}
```

### Lookup Room

```bash
GET /rooms/ABCD-1234

# Response
{
  "code": "ABCD-1234",
  "host": "192.168.1.100:7777",
  "name": "My Game",
  "players": 2,
  "max_players": 4
}
```

### Delete Room

```bash
DELETE /rooms/ABCD-1234
```

## Deployment

This service is stateless (in-memory store) by default. For production:
- Deploy behind HTTPS (Cloudflare, nginx, etc.)
- Add Redis backend for persistence
- Add rate limiting

Can run cheaply on:
- Cloudflare Workers
- Fly.io
- Railway
- Any small VPS
