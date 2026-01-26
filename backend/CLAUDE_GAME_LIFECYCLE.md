# Game Lifecycle (Authoritative Design)

## Purpose

Define how lobbies transition into games and back.
This document is authoritative and should be followed by all implementations.

## Core Principles

- WebSocket layer is authoritative for game flow
- Lobby "ready" is a signal, not persisted domain state
- Domain models do not track readiness yet
- No speculative refactors

## Lobby Phase

- Players join a lobby via WS
- Lobby becomes "full" at 2 players
- Players may send `set_ready` signals

## Ready Semantics

- Ready is ephemeral and session-scoped
- Ready state:
  - Is not persisted
  - Is cleared on disconnect
  - Is cleared on game start

## Game Start Conditions

- Exactly 2 connected players
- Both players have sent `set_ready`
- Server emits:
  - `game_starting`
  - `game_started`

## Out of Scope (Intentional)

- Persistent ready state
- Reconnect-in-progress
- Turn logic
- Game rules
