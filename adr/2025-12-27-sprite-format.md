# Sprite and Asset Format

**Status:** Accepted

## Context

Need a standardized format for sprites and visual assets. Format must be:
- Human-readable (for debugging and manual editing)
- Machine-generatable (AI tools like Google Nano, GPT, etc.)
- Simple enough that the format itself can be described in a prompt

Since we render to terminal, sprites are not traditional pixel art but ASCII/block representations.

## Options Considered

### PNG/Image Files

Traditional sprite sheets as images.

**Workflow:** Create in image editor → convert to ASCII at load time

**Pros:**
- Standard format, many tools
- Can use existing pixel art

**Cons:**
- Conversion to ASCII loses information
- Not directly editable as text
- AI image generation is separate from text generation
- Overkill for terminal rendering

### Custom Binary Format

Packed binary with palette and character data.

**Pros:**
- Compact file size
- Fast loading

**Cons:**
- Not human-readable
- Hard to generate with AI
- Custom tooling required
- Debugging nightmare

### JSON Sprite Definition

JSON object with character grid and metadata.

**Workflow:** Write/generate JSON → load directly

**Pros:**
- Human-readable
- AI can generate valid JSON easily
- Standard parsing libraries

**Cons:**
- Verbose for large sprites
- Escape characters for special chars
- No comments

### YAML Sprite Definition

YAML with character grid using block scalars.

**Workflow:** Write/generate YAML → load directly

**Pros:**
- Very human-readable
- Block scalars perfect for ASCII art
- Comments allowed
- AI generates YAML well

**Cons:**
- Whitespace sensitive (can cause bugs)
- Slightly slower parsing than JSON

### Plain Text with Header

Simple text file: metadata lines then raw ASCII.

**Pros:**
- Maximum simplicity
- Direct copy-paste of ASCII art
- Easy AI generation

**Cons:**
- Custom parser needed
- No standard structure
- Limited metadata support

## Decision

**YAML with block scalars for sprites:**

```yaml
# player.yaml
name: player_idle
width: 5
height: 4
anchor: [2, 3]  # Origin point (feet)
frames: 1

art: |
   O
  /|\
  / \

colors: |
  .W.
  WWW
  B.B

palette:
  W: "#FFFFFF"  # White
  B: "#4444FF"  # Blue
  .: transparent
```

For animations:
```yaml
name: player_walk
width: 5
height: 4
anchor: [2, 3]
frames: 2
framerate: 8  # FPS

art:
  - |
     O
    /|\
    / \
  - |
     O
    /|\
     |\\

colors:
  - |
    .W.
    WWW
    B.B
  - |
    .W.
    WWW
    B.B
```

### AI Generation Prompt Template

```
Generate a YAML sprite for [description].
Use this format:
- name: identifier
- width/height: dimensions in characters
- art: block scalar with ASCII art using these chars: O o @ # = - | / \ _ . space
- colors: matching grid with single-char palette keys
- palette: map of keys to hex colors, use "transparent" for empty

Example:
name: enemy_slime
width: 3
height: 2
art: |
  ___
 (o_o)
colors: |
  GGG
  GWGW
palette:
  G: "#00FF00"
  W: "#FFFFFF"
```

### Directory Structure

```
assets/
  sprites/
    player/
      idle.yaml
      walk.yaml
      jump.yaml
      attack.yaml
    enemies/
      slime.yaml
      bat.yaml
    items/
      coin.yaml
      heart.yaml
    tiles/
      ground.yaml
      platform.yaml
```

## Consequences

- **AI-friendly** - YAML is well-understood by LLMs, can generate sprites from descriptions
- **Human-editable** - Can tweak sprites in any text editor
- **Git-friendly** - Text diffs show exactly what changed
- **Whitespace sensitivity** - Must be careful with trailing spaces in art blocks
- **Validation needed** - Should validate width/height match actual art dimensions
- **Color layer optional** - Can omit colors block for single-color sprites
