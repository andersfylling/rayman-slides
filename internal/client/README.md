# client

Game client. Handles terminal rendering, input capture, and server communication.

## Responsibilities

1. Detect terminal capabilities
2. Capture keyboard input
3. Connect to server (or start embedded server)
4. Send inputs each tick
5. Receive state snapshots
6. Interpolate and render

## Usage

```go
cfg := client.Config{
    ServerAddr: "localhost:7777",  // Empty for local play
    PlayerName: "Player1",
    RenderMode: client.RenderAuto,
}

c := client.New(cfg)
if err := c.Connect(); err != nil {
    log.Fatal(err)
}
c.Run()  // Blocks until quit
```

## Render Modes

| Mode | Description |
|------|-------------|
| `RenderAuto` | Detect best mode for terminal |
| `RenderASCII` | Plain ASCII characters |
| `RenderHalfBlock` | Unicode half-blocks with color |
| `RenderBraille` | Braille patterns (highest res) |

## Local Play

When `ServerAddr` is empty, client starts an embedded server automatically. This provides identical gameplay to multiplayer but without network latency.
