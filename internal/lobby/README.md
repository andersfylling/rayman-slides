# lobby

Game discovery and room codes.

## Room Codes

Short codes like `ABCD-1234` that map to server addresses. Easier to share than IP:port.

```go
store := lobby.NewRoomStore(4 * time.Hour)

// Host creates room
room, _ := store.Create("192.168.1.100:7777", "My Game", 4)
fmt.Println(room.Code)  // "ABCD-1234"

// Player looks up room
room, _ := store.Lookup("ABCD-1234")
fmt.Println(room.Host)  // "192.168.1.100:7777"
```

## Code Format

`XXXX-XXXX` using charset `ABCDEFGHJKLMNPQRSTUVWXYZ23456789`

Ambiguous characters excluded: I, O, 0, 1

## Server Integration

The dedicated server can register with the lookup service:

```bash
./rayserver --register --lookup https://lookup.example.com
```

This creates a room and prints the code. When the server shuts down, it deletes the room.

## Direct Connect

Room codes are optional. Players can always connect directly via IP:port if they prefer.

See `adr/2025-12-27-lobby-system.md`.
