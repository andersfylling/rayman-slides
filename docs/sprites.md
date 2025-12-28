# Sprite System

This document defines all sprites needed for the game and their specifications.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            SPRITE PIPELINE                                  │
└─────────────────────────────────────────────────────────────────────────────┘

  ECS Component              Renderer                    Display
  ─────────────              ────────                    ───────
       │                         │                          │
  ┌─────────┐               ┌─────────┐               ┌─────────┐
  │ Sprite  │               │  Atlas  │               │  GPU    │
  │         │──SpriteID────▶│         │──Texture─────▶│  Quad   │
  │ ID: str │               │ Lookup  │               │         │
  │ Color   │               └─────────┘               └─────────┘
  └─────────┘
       │
       │ Updated by game logic:
       │  - Attack system (charge/punch states)
       │  - Movement system (walk/jump/idle)
       │  - Damage system (hurt/death)
       │
```

## Sprite ID Format

Sprite IDs follow a hierarchical naming convention:

```
{entity}_{state}_{variant}

Examples:
  player                    # Default idle
  player_walk_1             # Walk animation frame 1
  player_charge_right_2     # Charging attack, facing right, pulse frame 2
  slime_idle                # Slime default
  tile_ground_grass         # Ground tile with grass top
```

## Required Sprites

### Player (Rayman-style, limbless)

The player is a limbless character inspired by Rayman. Body parts float independently.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PLAYER ANATOMY                                                             │
│                                                                             │
│        ┌───┐                                                                │
│        │ O │  ← Head (floats)                                               │
│        └───┘                                                                │
│                                                                             │
│     ○       ○  ← Hands (float, no arms)                                     │
│                                                                             │
│        ┌───┐                                                                │
│        │   │  ← Torso                                                       │
│        └───┘                                                                │
│                                                                             │
│     ○       ○  ← Feet (float, no legs)                                      │
│                                                                             │
│  Anchor point: Bottom center of torso (feet level)                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `player_idle` | Standing still | 1-4 (breathing) | 32x48 |
| `player_walk_1..4` | Walking cycle | 4 | 32x48 |
| `player_jump` | Jumping/falling | 1 | 32x48 |
| `player_charge_right_1..3` | Charging attack, right | 3 (pulsing glow) | 32x48 |
| `player_charge_left_1..3` | Charging attack, left | 3 (pulsing glow) | 32x48 |
| `player_punch_right` | Punch pose, right | 1 | 48x48 |
| `player_punch_left` | Punch pose, left | 1 | 48x48 |
| `player_hurt` | Taking damage | 1-2 | 32x48 |
| `player_helicopter` | Helicopter hair (future) | 2-4 | 32x56 |
| `player_grapple` | Grappling (future) | 1 | 32x48 |

**Animation Notes:**
- Idle: Subtle breathing animation (torso bobs slightly)
- Walk: Feet alternate, hands swing opposite to feet
- Jump: Arms up, feet tucked
- Charge: Fist glows with increasing intensity (3 pulse frames, cycle at 10 TPS)
- Punch: Arm extended, fist at end of arm line

### Fist (Projectile)

The flying fist is Rayman's signature attack.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  FIST PROJECTILE                                                            │
│                                                                             │
│     ───○))   Facing right                                                   │
│                                                                             │
│     ((○───   Facing left                                                    │
│                                                                             │
│  Trail effect: Motion blur or speed lines                                   │
│  Anchor point: Center of fist                                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `fist_right` | Flying fist, right | 1-2 (spin) | 16x16 |
| `fist_left` | Flying fist, left | 1-2 (spin) | 16x16 |
| `fist_impact` | Hit effect | 3-4 | 24x24 |

### Enemies

#### Slime

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  SLIME                                                                      │
│                                                                             │
│      ┌───┐                                                                  │
│     /  ● ●\   ← Eyes                                                        │
│    │       │                                                                │
│     \_____/   ← Jiggly blob body                                            │
│                                                                             │
│  Anchor point: Bottom center                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `slime_idle` | Idle jiggle | 2-4 | 24x20 |
| `slime_move_1..2` | Hopping movement | 2 | 24x24 |
| `slime_hurt` | Taking damage | 1 | 24x20 |
| `slime_death` | Death splat | 3-4 | 32x16 |

#### Bat

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  BAT                                                                        │
│                                                                             │
│     \   ^   /                                                               │
│      \ ●_● /   ← Face with fangs                                            │
│       \   /                                                                 │
│        \_/     ← Wings flap                                                 │
│                                                                             │
│  Anchor point: Center                                                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `bat_fly_1..4` | Flying cycle | 4 | 24x24 |
| `bat_hurt` | Taking damage | 1 | 24x24 |
| `bat_death` | Death poof | 3 | 24x24 |

### Tiles (16x16 base grid)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  TILE TYPES                                                                 │
│                                                                             │
│  ████████  SOLID      ════════  PLATFORM    ~~~~~~~~  WATER                 │
│  ████████  (blocks)   ════════  (one-way)   ~~~~~~~~  (hazard)              │
│                                                                             │
│  ╔══════╗  CAGE       ▲▲▲▲▲▲▲▲  SPIKES      ╫╫╫╫╫╫╫╫  LADDER                │
│  ║ ○○○○ ║  (rescue)   ▲▲▲▲▲▲▲▲  (hazard)    ╫╫╫╫╫╫╫╫  (climb)               │
│  ╚══════╝                                                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Sprite ID | Description | Variants |
|-----------|-------------|----------|
| `tile_ground` | Solid ground | `_grass`, `_dirt`, `_stone` |
| `tile_platform` | One-way platform | `_wood`, `_stone`, `_cloud` |
| `tile_wall` | Vertical wall | `_left`, `_right`, `_both` |
| `tile_corner` | Corner pieces | `_tl`, `_tr`, `_bl`, `_br` |
| `tile_water` | Water surface | `_top`, `_body` (animated) |
| `tile_spikes` | Damage spikes | `_up`, `_down`, `_left`, `_right` |
| `tile_ladder` | Climbable | `_top`, `_middle`, `_bottom` |
| `tile_cage` | Rescue cage | `_closed`, `_open` |

### Collectibles

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `ting` | Yellow orb (currency) | 4 (spin) | 12x12 |
| `ting_collect` | Collection effect | 4 | 16x16 |
| `health_small` | Small health pickup | 2 (pulse) | 12x12 |
| `health_big` | Large health pickup | 2 (pulse) | 16x16 |
| `1up` | Extra life | 2 (pulse) | 16x16 |

### Effects

| Sprite ID | Description | Frames | Size (px) |
|-----------|-------------|--------|-----------|
| `dust_land` | Landing dust | 4 | 16x8 |
| `dust_run` | Running dust | 3 | 12x8 |
| `splash` | Water splash | 4 | 24x16 |
| `sparkle` | Generic sparkle | 4 | 8x8 |

## Color Palette

The game uses a whimsical, fairy-tale aesthetic with saturated colors.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  PRIMARY PALETTE                                                            │
│                                                                             │
│  Player:     #4CAF50 (green)     #8BC34A (light green)                      │
│  Fist:       #FFEB3B (yellow)    #FFC107 (orange glow)                      │
│  Slime:      #4CAF50 (green)     #81C784 (light green)                      │
│  Bat:        #9C27B0 (purple)    #CE93D8 (light purple)                     │
│                                                                             │
│  Ground:     #795548 (brown)     #8D6E63 (light brown)                      │
│  Grass:      #388E3C (green)     #66BB6A (light green)                      │
│  Water:      #2196F3 (blue)      #64B5F6 (light blue)                       │
│  Sky:        #1A237E (dark)      #3F51B5 (indigo)                           │
│                                                                             │
│  UI:         #FFFFFF (white)     #FFC107 (gold accents)                     │
│  Damage:     #F44336 (red flash)                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Sprite Atlas Layout

Sprites are packed into a single texture atlas for efficient GPU rendering.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ATLAS ORGANIZATION (512x512 or 1024x1024)                                  │
│                                                                             │
│  ┌────────────────────────────────────────────┐                             │
│  │  PLAYER SPRITES (row 0-2)                  │                             │
│  │  idle | walk1-4 | jump | charge | punch    │                             │
│  ├────────────────────────────────────────────┤                             │
│  │  ENEMY SPRITES (row 3-4)                   │                             │
│  │  slime_* | bat_*                           │                             │
│  ├────────────────────────────────────────────┤                             │
│  │  TILES (row 5-8)                           │                             │
│  │  16x16 grid of all tile variants           │                             │
│  ├────────────────────────────────────────────┤                             │
│  │  EFFECTS & COLLECTIBLES (row 9-10)         │                             │
│  │  fist | ting | health | dust | splash      │                             │
│  └────────────────────────────────────────────┘                             │
│                                                                             │
│  Atlas metadata stored in sprites.json:                                     │
│  {                                                                          │
│    "player_idle": { "x": 0, "y": 0, "w": 32, "h": 48, "anchor": [16,48] },  │
│    "player_walk_1": { "x": 32, "y": 0, "w": 32, "h": 48, "anchor": [16,48] }│
│    ...                                                                      │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Animation System

Animations are handled by cycling sprite IDs at fixed intervals:

```go
// Animation definition
type Animation struct {
    Frames   []string      // Sprite IDs to cycle
    Duration time.Duration // Time per frame
    Loop     bool          // Restart after last frame
}

// Example animations
var Animations = map[string]Animation{
    "player_walk": {
        Frames:   []string{"player_walk_1", "player_walk_2", "player_walk_3", "player_walk_4"},
        Duration: 100 * time.Millisecond,
        Loop:     true,
    },
    "player_charge": {
        Frames:   []string{"player_charge_right_1", "player_charge_right_2", "player_charge_right_3"},
        Duration: 100 * time.Millisecond,
        Loop:     true,
    },
}
```

## Implementation Phases

### Phase 1: Placeholder Sprites (Current)
- Colored rectangles in Gio renderer
- SpriteID maps to hard-coded colors
- No animation

### Phase 2: Static Sprites
- Load PNG atlas with `image/png`
- Implement atlas lookup by SpriteID
- Draw textured quads in Gio
- Single frame per entity

### Phase 3: Animated Sprites
- Add `AnimationState` component to ECS
- Tick-based frame advancement
- Support for one-shot and looping animations

### Phase 4: Effects & Polish
- Particle system for dust, sparkles
- Screen shake on damage
- Smooth sprite interpolation between ticks

## File Structure

```
assets/
  sprites/
    atlas.png           # Combined sprite sheet
    sprites.json        # Atlas metadata (positions, anchors)
  profiles/
    default/            # Default visual theme
      atlas.png
      sprites.json
    retro/              # Alternative 8-bit style
      atlas.png
      sprites.json
```

## Current Sprite ID Usage

From `internal/game/world.go` and `internal/render/gio.go`:

| ID Pattern | Used By | Current Rendering |
|------------|---------|-------------------|
| `player` | Player idle | Green rectangle |
| `player_charge_right_N` | Charging (right) | Yellow rectangle |
| `player_charge_left_N` | Charging (left) | Yellow rectangle |
| `player_punch_right` | Punching (right) | Light green rectangle |
| `player_punch_left` | Punching (left) | Light green rectangle |
| `fist_right` | Flying fist (right) | Yellow small rectangle |
| `fist_left` | Flying fist (left) | Yellow small rectangle |
| `slime` | Slime enemy | Green rectangle |
| `bat` | Bat enemy | Purple rectangle |
