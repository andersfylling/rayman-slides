# Configuration File Format

**Status:** Accepted

## Context

Need a configuration file for game settings. Users should be able to manually edit it if the in-game menu has bugs or doesn't expose certain options.

## Options Considered

### JSON

Standard, widely supported.

**Pros:**
- Universal support
- Strict parsing (catches errors)
- Fast parsing

**Cons:**
- No comments (users can't document their changes)
- Verbose (lots of quotes and braces)
- Trailing comma errors
- Not great for manual editing

### TOML

Designed for config files (Cargo.toml, pyproject.toml).

**Pros:**
- Clean syntax
- Comments supported
- Good for flat/shallow configs
- Familiar to developers

**Cons:**
- Nested structures get awkward
- Less familiar to non-developers
- Fewer libraries than JSON/YAML

### YAML

Human-readable data serialization.

**Pros:**
- Very readable
- Comments supported
- Good for nested structures
- Familiar format (docker-compose, k8s, etc.)
- AI tools understand it well

**Cons:**
- Whitespace sensitive
- "Norway problem" (bare `no` becomes boolean)
- Multiple ways to express same thing

### INI

Classic config format.

**Pros:**
- Dead simple
- Comments supported
- Very readable for simple configs

**Cons:**
- No nested structures
- No arrays/lists
- No standardized parsing
- Too limited for game config

### Custom Format

Design our own syntax.

**Pros:**
- Exactly what we need

**Cons:**
- Learning curve for users
- Must write parser
- No tooling support
- Bad idea

## Decision

**YAML for configuration:**

Rationale:
- Matches sprite format (consistency)
- Comments let users document their tweaks
- Nested structures for keybindings, video settings, etc.
- Users familiar with it from other tools

### Config File Location

```
Linux:   ~/.config/rayman/config.yaml
macOS:   ~/Library/Application Support/rayman/config.yaml
Windows: %APPDATA%\rayman\config.yaml
```

Fallback: `./config.yaml` in game directory.

### Config Structure

```yaml
# Rayman Terminal Configuration
# Edit this file to customize game settings
# Delete this file to reset to defaults

version: 1  # Config schema version

# Video/rendering settings
video:
  render_mode: auto  # auto, ascii, halfblock, braille
  fps_limit: 60
  show_fps: false
  fullscreen: false

# Audio settings (future)
audio:
  master_volume: 100
  music_volume: 80
  sfx_volume: 100
  mute: false

# Gameplay settings
gameplay:
  difficulty: normal  # easy, normal, hard
  screenshake: true
  show_damage_numbers: true

# Controls - action: key
# Use lowercase letters, or special keys: space, enter, escape, up, down, left, right, tab
controls:
  move_left: a
  move_right: d
  jump: space
  attack: j
  use: k
  pause: escape

  # Alternative bindings (optional)
  alt_move_left: left
  alt_move_right: right
  alt_jump: w

# Network settings
network:
  player_name: "Player"
  default_port: 7777
  lookup_server: "https://lookup.rayman.example.com"

# Advanced settings (change at your own risk)
advanced:
  tick_rate: 60
  interpolation_buffer: 3  # Snapshots to buffer
  debug_mode: false
  log_level: info  # debug, info, warn, error
```

### Loading Priority

1. Command-line flags (highest priority)
2. Environment variables (RAYMAN_*)
3. Config file
4. Built-in defaults (lowest priority)

### Schema Versioning

Config has `version` field. On load:
- If version < current: migrate and save
- If version > current: warn user, use defaults for unknown fields
- If missing: assume version 1

### Validation

```go
type Config struct {
    Version  int           `yaml:"version"`
    Video    VideoConfig   `yaml:"video"`
    Audio    AudioConfig   `yaml:"audio"`
    Gameplay GameplayConfig `yaml:"gameplay"`
    Controls ControlsConfig `yaml:"controls"`
    Network  NetworkConfig  `yaml:"network"`
    Advanced AdvancedConfig `yaml:"advanced"`
}

func (c *Config) Validate() error {
    // Check render_mode is valid enum
    // Check volumes are 0-100
    // Check controls are valid key names
    // etc.
}
```

### Default Config Generation

On first run or if config missing:
1. Create config directory
2. Write default config with comments
3. Log location to user

## Consequences

- **User-editable** - Users can fix settings if menu breaks
- **Comments preserved** - Users can document their tweaks
- **Version migration** - Can evolve config over time
- **Consistency** - Same format as sprites
- **Whitespace bugs** - YAML indentation errors possible
- **Validation required** - Must validate on load, provide helpful errors
- **Quote strings** - Player names and paths should be quoted to avoid YAML gotchas
