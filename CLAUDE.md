# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

See [AGENTS.md](./AGENTS.md) for project-specific instructions that apply to all AI coding agents.

## About AGENTS.md

[AGENTS.md](https://agents.md/) is a simple, open format for guiding AI coding agents. It provides a standardized way to give agents project-specific context, build commands, code style guidelines, and workflows. The format is supported by many AI coding tools including Claude Code, GitHub Copilot, Cursor, OpenAI Codex, and others.

Key features:
- Standard Markdown with no required schema
- Hierarchical support for monorepos (agents read the nearest AGENTS.md in the directory tree)
- Complements README files by targeting agent-specific needs rather than human contributors
