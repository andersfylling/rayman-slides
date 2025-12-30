# Sprite Assets

This directory contains sprite profiles for the game. Each profile is a self-contained folder with a sprite atlas and metadata.

## Directory Structure

```
assets/sprites/
├── README.md           # This file
├── default/            # Default sprite profile
│   ├── atlas.jpg       # Sprite sheet image (1408x768)
│   └── atlas.json      # Sprite region definitions
└── <profile>/          # Additional profiles
    ├── atlas.jpg
    └── atlas.json
```

## atlas.json Format

```json
{
  "image": "atlas.jpg",
  "sprites": {
    "sprite_name": {
      "x": 0,              // X position in atlas (pixels)
      "y": 0,              // Y position in atlas (pixels)
      "w": 100,            // Width (pixels)
      "h": 100,            // Height (pixels)
      "anchorX": 50,       // Anchor X relative to sprite top-left
      "anchorY": 100,      // Anchor Y relative to sprite top-left
      "hitX": 10,          // Hitbox X offset (optional)
      "hitY": 0,           // Hitbox Y offset (optional)
      "hitW": 80,          // Hitbox width (optional)
      "hitH": 100          // Hitbox height (optional)
    }
  }
}
```

## Anchor Points

The anchor point defines the sprite's origin for positioning:
- For characters: typically at feet (anchorX=width/2, anchorY=height)
- For tiles: typically at top-left (anchorX=0, anchorY=0) or bottom (anchorY=height)
- For centered objects: (anchorX=width/2, anchorY=height/2)

## Hitboxes

Hitbox fields are optional. When omitted, the hitbox equals the sprite bounds.
Hitbox coordinates are relative to the sprite's top-left corner.

## Sprite Naming Convention

| Prefix | Type | Example |
|--------|------|---------|
| `player_` | Player animations | `player_idle`, `player_walk_1` |
| `blob_` | Blob enemy | `blob_1`, `blob_jump_1` |
| `bat_` | Bat enemy | `bat_1`, `bat_death` |
| `tile_` | Terrain tiles | `tile_grass`, `tile_water` |
| `fist_` | Fist projectile | `fist_1`, `fist_2` |
| `orb_` | Collectible orbs | `orb_1`, `orb_2` |
| `smoke_` | Smoke effects | `smoke_1`, `smoke_4` |
| `cage_` | Cage objects | `cage_closed`, `cage_open` |

## Animation Sequences

| Animation | Sprites | Loop Pattern |
|-----------|---------|--------------|
| Player walk | `player_walk_1` → `player_walk_4` | 1-2-3-4-1... |
| Player attack | `player_attack_1` → `player_attack_2` | 1-2-1... |
| Bat fly | `bat_1` → `bat_5` | 1-2-3-4-5-4-3-2-1... |
| Blob idle | `blob_1` → `blob_2` | 1-2-1... |
| Orb shine | `orb_1` → `orb_3` | 1-2-3-2-1... (ping-pong) |
| Smoke grow | `smoke_1` → `smoke_4` | 1-2-3-4 (once) |

## Tools

### Sprite Editor
```bash
make sprite-editor
```
Interactive GUI for editing sprite regions:
- Draw boxes with left-drag
- Select/move with click + arrow keys
- Mode switch: `1`=Box, `2`=Anchor, `3`=Hitbox
- Save with `S`

### Sprite Debug
```bash
make sprites-debug
```
Generates visualization files:
- `sprites.debug.png` - Atlas overlay with borders and labels
- `sprites.debug.gif` - Static grid of all sprites
- `sprites.animated.gif` - Animated preview

## Creating a New Profile

1. Create a new folder: `assets/sprites/myprofile/`
2. Add `atlas.jpg` (sprite sheet image)
3. Add `atlas.json` (copy from default, edit regions)
4. Update game config to use `"sprite_profile": "myprofile"`

## Tile Dimensions

All tiles (except cloud) should be 108x108 pixels for consistent grid alignment.
