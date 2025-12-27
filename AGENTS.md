# AGENTS.md

## Project Overview

An open-source side-scrolling platformer inspired by Rayman 1 (1995). This is a fan recreation using entirely custom/AI-generated assets—no original game data is used.

## Game Concept (Based on Rayman 1)

### Core Mechanics
- 2D side-scrolling platformer
- Player character with detachable limbs (no connecting arms/legs—originally a technical limitation that became iconic)
- Progressive ability unlocks from a fairy NPC:
  1. **Telescopic fist** - punch enemies and break cages
  2. **Hanging** - grab onto ledges
  3. **Grappling fist** - swing from flying rings
  4. **Helicopter** - use hair to glide/hover
  5. **Running** - move faster, longer jumps (replaces grimace ability)
- Temporary powers: magic seeds (grow platforms), super helicopter (free flight), shrinking (access small areas)
- Collectibles: blue orbs ("Tings"), 100 = extra life
- 102 caged creatures (6 per level) scattered across worlds—ALL must be freed to unlock final world
- Fist power-ups: Speed Fist (faster/farther), Golden Fist (2x damage)
- Health: 3-5 hit points, restored by Power items
- Lives system with checkpoints (Photographer character)

### World Structure (6 Thematic Worlds)
Each world has 3-4 levels plus a boss. Half are "natural," half are "imaginary":

1. **The Dream Forest** (natural) - lush jungle with pink plants, lagoons, swamps
   - Boss: Moskito (giant mosquito)
   - Mid-boss: Bzzit (friendly mosquito after defeat, becomes rideable)

2. **Band Land** (imaginary) - musical instruments and sheet music landscape
   - Boss: Mr Sax (walking saxophone, shoots explosive notes)
   - Features: drum platforms, slippery sheet music, cloud areas

3. **Blue Mountains** (natural) - cold mountain range
   - Boss: Mr Stone (rock creature)
   - NPC: The Musician (gives super helicopter power)

4. **Picture City** (imaginary) - artwork and art supplies
   - Boss: Space Mama (large opera singer with rolling pin weapons)
   - Features: eraser plains, pencil obstacles

5. **The Caves of Skops** (natural) - dark underground caverns
   - Boss: Mr Skops (giant scorpion with homing laser tail)
   - NPC: Joe the Extra-Terrestrial (alien restaurant owner)
   - Mechanic: firefly lighting in pitch-dark sections

6. **Candy Château** (imaginary) - sweets and crockery, final world
   - Boss: Mr Dark (main villain) + hybrid boss forms
   - Hazards: Bad Rayman (evil clone shadow), control-reversing spells, forced running, ability removal

### Key NPCs
- **The Magician** - narrator, offers bonus stages for 10 Tings
- **Betilla the Fairy** - grants permanent powers throughout the game
- **The Photographer** - acts as checkpoint

### Final Boss Mechanics
Mr Dark's battle includes:
- Stealing player's fist ability
- Fire wall traps
- Transforming into hybrid creatures combining previous bosses
- Player shrinking spell

### Visual Style
- Colorful, whimsical 2D art inspired by Celtic, Chinese, and Russian fairy tales
- Smooth animations (target 60 FPS)
- Innocent/childish presentation contrasting with challenging gameplay
- Parallax scrolling backgrounds

### Difficulty
The original was notoriously difficult (never play-tested). Our version should:
- Offer casual/standard difficulty options
- Provide adequate checkpoints
- Balance challenge with accessibility

## Asset Strategy ("Profiles")

All game assets are custom-generated to keep the project fully open source:
- **Character skins**: AI-generated sprites reminiscent of the limbless design
- **Environment tiles**: Original artwork inspired by the fairy-tale aesthetic
- **Music/SFX**: Original whimsical compositions
- **No copyrighted assets**: Everything is created fresh

"Profiles" = interchangeable skin/asset packs that can be swapped to change the game's look while maintaining the same mechanics.

## Technology Stack

Go

### Rendering Backends

The game supports multiple rendering backends via a pluggable architecture:

1. **Terminal** (primary) - ASCII/Unicode rendering for universal Linux compatibility, no GPU drivers required
2. **CPU Software Renderer** - Pure CPU rendering for environments without GPU acceleration
3. **Vulkan** - Hardware-accelerated graphics for full visual fidelity

This allows the same game logic to run anywhere from a headless server to a gaming PC.

## Build Commands

*To be added as the project develops.*

## Test Commands

*To be added as the project develops.*

## Code Style

*To be added as the project develops.*
