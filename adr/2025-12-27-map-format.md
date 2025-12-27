# Map Format and Generation

**Status:** Accepted

## Context

Need a map system for a platformer. Some players/modders want hand-crafted levels, others want procedural generation. Must support both use cases.

## Options Considered

### Tiled JSON (.tmj)

Industry standard tile map editor. Exports to JSON. Supports layers, objects, tilesets.

**Workflow:** Download Tiled → create tileset → paint map → export JSON → load in game

**Pros:**
- Visual editor with undo, layers, copy/paste
- Large community, tutorials, existing tilesets
- Object layers for spawn points, triggers, secrets

**Cons:**
- External tool dependency
- Schema tied to Tiled's format (verbose)
- Players need to install Tiled to create maps

### Custom JSON

Define our own simple schema. Edit by hand or build a simple editor.

**Workflow:** Write JSON in text editor or build in-game editor → load directly

**Pros:**
- Full control over schema
- No external dependencies
- Can optimize for our specific needs

**Cons:**
- No visual editor (unless we build one)
- Hand-editing coordinates is tedious and error-prone
- Reinventing the wheel

### Binary Format

Custom binary format. Fast parsing, small files.

**Workflow:** Build custom editor → save binary → load in game

**Pros:**
- Fastest load times
- Smallest file size
- Can embed directly in executable

**Cons:**
- Not human-readable (hard to debug)
- Must build tooling from scratch
- Versioning/migration is painful

### Procedural Generation

Generate maps at runtime from seeds or rules.

**Workflow:** Define generation rules → seed produces deterministic map

**Pros:**
- Infinite variety
- No asset files to distribute
- Seed sharing for multiplayer

**Cons:**
- Hard to ensure quality/balance
- Precision platforming needs hand-tuned jumps
- "Samey" feel without extensive rule tuning

### Hybrid: Chunks + Assembly

Hand-craft room/chunk templates. Procedural system assembles them.

**Workflow:** Create room templates (any format) → define connection rules → generator assembles

**Pros:**
- Hand-crafted feel with variety
- Each chunk is tested/balanced
- Supports both use cases

**Cons:**
- Need many chunks to avoid repetition
- Connection logic adds complexity
- Still need a format for chunks

## Decision

**Hybrid approach with Tiled as primary format:**

1. **Tiled JSON for chunks/rooms** - Visual editing, community familiarity
2. **Simple internal representation** - Convert Tiled on load to our own struct
3. **Procedural assembler** - Optional system to combine chunks
4. **Seed-based generation** - Deterministic for multiplayer sync

```
maps/
  chunks/          # Tiled .tmj files (hand-crafted rooms)
  campaigns/       # Full hand-crafted levels
  tilesets/        # Shared tileset definitions
```

Players can:
- Play hand-crafted campaign levels
- Play procedurally assembled runs
- Create custom maps with Tiled
- Share seeds for procedural runs

## Consequences

- **Two workflows** - Hand-crafted (Tiled) and procedural (seeds) both supported
- **Tiled dependency** - Map creators need external tool, but it's free and popular
- **Conversion layer** - Must parse Tiled format → internal format (one-time cost)
- **Chunk design** - Need to define connection points/rules for procedural assembly
- **Determinism** - Procedural gen must be deterministic for multiplayer
