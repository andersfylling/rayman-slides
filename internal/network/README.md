# network

Transport layer for client-server communication.

## Architecture

```
Client                          Server
  │                               │
  │──── TCP Connection ───────────│
  │                               │
  │ InputFrame ──────────────────►│
  │                               │
  │◄────────────── StateSnapshot  │
  │                               │
```

## Usage

### Server

```go
transport := network.NewTCPTransport()
transport.Listen(":7777")

for {
    conn, _ := transport.Accept()
    go handleClient(conn)
}
```

### Client

```go
transport := network.NewTCPTransport()
transport.Connect("localhost:7777")
```

## Message Framing

TCP is a stream, so we length-prefix messages:

```
[4 bytes: length][N bytes: payload]
```

## Future: QUIC

TCP works but has head-of-line blocking. QUIC upgrade planned if latency becomes an issue. The `Transport` interface allows swapping implementations.

See `adr/2025-12-27-network-architecture.md`.
