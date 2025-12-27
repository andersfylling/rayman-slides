# User Personas

These personas represent different types of users and contributors. Use them to guide feature development and ensure the project serves diverse needs.

## Players

### The Nostalgic Recreator

> "I want to manually create maps that look similar to the original Rayman, maybe with slight changes or fixes."

**Goals:**
- Recreate classic Rayman 1 levels faithfully
- Fix frustrating sections from the original
- Share levels with others who remember the game

**Needs:**
- [ ] Visual map editor or easy-to-edit format
- [ ] Tileset that matches original aesthetic
- [ ] Ability to place enemies, collectibles, checkpoints
- [ ] Import/reference original level layouts
- [ ] Export levels to share with community

**Relevant ADRs:** [Map Format](../adr/2025-12-27-map-format.md)

---

### The Roguelike Explorer

> "I want pseudo-random maps so every playthrough feels fresh, like a new adventure."

**Goals:**
- Endless replayability
- Surprise and discovery each session
- Shareable seeds for interesting runs

**Needs:**
- [ ] Procedural level generator with configurable difficulty
- [ ] Seed system for reproducible generation
- [ ] Biome/theme variety (forest, caves, etc.)
- [ ] Guaranteed solvability (no impossible sections)
- [ ] Daily/weekly challenge seeds

**Relevant ADRs:** [Map Format](../adr/2025-12-27-map-format.md)

---

### The Speedrunner

> "I want consistent, optimized maps where I can practice and shave milliseconds off my time."

**Goals:**
- Master precise movement
- Compete on leaderboards
- Find and exploit optimal routes

**Needs:**
- [ ] Deterministic physics (tick-based, no floating point drift)
- [ ] Frame-perfect input recording/playback
- [ ] Ghost replays to race against
- [ ] Built-in timer with splits
- [ ] Leaderboard integration

**Relevant ADRs:** [Game Loop](../adr/2025-12-27-game-loop-tick-based.md)

---

### The Casual Couch Player

> "I just want to pick up and play without reading docs or configuring anything."

**Goals:**
- Immediate fun with minimal setup
- Clear progression and goals
- Not punishing difficulty

**Needs:**
- [ ] Works out of the box (`./rayman` and play)
- [ ] Built-in tutorial or intuitive controls
- [ ] Difficulty options (casual/normal/hard)
- [ ] Save system with checkpoints
- [ ] Controller support (eventually)

---

### The Party Host

> "I want to play with friends on my local network or online."

**Goals:**
- Easy multiplayer setup
- Fun co-op or competitive modes
- Low-latency gameplay

**Needs:**
- [ ] Room codes for easy joining
- [ ] Embedded server for host-and-play
- [ ] Dedicated server for persistent games
- [ ] Co-op mode (shared or split objectives)
- [ ] Versus mode (race, battle, etc.)

**Relevant ADRs:** [Network](../adr/2025-12-27-network-architecture.md), [Lobby](../adr/2025-12-27-lobby-system.md)

---

## Creators

### The Level Designer

> "I want powerful tools to create elaborate custom levels with triggers, secrets, and custom logic."

**Goals:**
- Express creative vision through level design
- Build complex, polished experiences
- Share creations with the community

**Needs:**
- [ ] Full-featured level editor (GUI or TUI)
- [ ] Trigger system (doors, switches, events)
- [ ] Custom enemy placement and patrol paths
- [ ] Secret areas and alternate routes
- [ ] Level testing mode (quick iteration)
- [ ] Level packaging and distribution

**Relevant ADRs:** [Map Format](../adr/2025-12-27-map-format.md)

---

### The Artist

> "I want to create custom sprites and visual themes (profiles) for the game."

**Goals:**
- Express visual creativity
- Create cohesive aesthetic themes
- Share art packs with the community

**Needs:**
- [ ] Clear sprite format specification
- [ ] PNG → terminal converter tool
- [ ] Preview tool to see sprites in-game
- [ ] Profile system for swappable asset packs
- [ ] Template/starter kit for new profiles

**Relevant ADRs:** [Sprite Format](../adr/2025-12-27-sprite-format.md)

---

### The Modder

> "I want to extend the game with new abilities, enemies, or mechanics without forking."

**Goals:**
- Add custom content without core changes
- Share mods with others
- Combine multiple mods

**Needs:**
- [ ] Plugin/mod loading system
- [ ] Lua or scripting for custom logic
- [ ] Hook points for game events
- [ ] Mod manager for enabling/disabling
- [ ] Documentation for mod API

---

## Technical

### The Server Operator

> "I want to run a public server for the community with proper administration."

**Goals:**
- Stable, long-running server
- Control over who can join
- Monitoring and logging

**Needs:**
- [ ] Headless server binary
- [ ] Config file for all settings
- [ ] Admin commands (kick, ban, etc.)
- [ ] Logging and metrics
- [ ] Docker image for easy deployment

**Relevant ADRs:** [Network](../adr/2025-12-27-network-architecture.md)

---

### The Contributor

> "I want to contribute code but I'm new to the project."

**Goals:**
- Understand codebase quickly
- Find meaningful work to do
- Get PRs merged smoothly

**Needs:**
- [x] Quick Start in README
- [x] CONTRIBUTING.md guide
- [x] Example tutorials (new enemy, etc.)
- [x] Issue templates
- [ ] Good first issues labeled
- [ ] Architecture overview diagram
- [ ] Code walkthrough video/doc

---

## Persona Priority Matrix

Which personas should we focus on first?

| Persona | Priority | Rationale |
|---------|----------|-----------|
| Casual Player | High | Core game must be fun first |
| Nostalgic Recreator | High | Primary target audience |
| Contributor | High | Need community to grow |
| Roguelike Explorer | Medium | Adds replayability |
| Party Host | Medium | Multiplayer is a key feature |
| Level Designer | Medium | Extends content |
| Artist | Medium | Enables profiles feature |
| Speedrunner | Low | Niche but important for polish |
| Modder | Low | Advanced feature, later |
| Server Operator | Low | After multiplayer works |

## Using Personas

When planning features or writing issues:

1. **Reference the persona:** "This helps the Roguelike Explorer by..."
2. **Check the needs list:** Are we addressing their requirements?
3. **Consider trade-offs:** Does this help one persona but hurt another?

When a feature is implemented, check off the relevant needs in this document.
