# Backend (Go / Gin) Instructions

This backend implements an authoritative game server for a real-time, turn-based multiplayer battle game.

## Project structure (expected)

backend/
├── cmd/
│ └── api/
│ └── main.go
├── internal/
│ ├── routes/ # Route registration & versioning
│ ├── controllers/ # HTTP / WS handlers (thin)
│ ├── services/ # Game orchestration & use-cases
│ ├── game/ # Core battle logic & domain models
│ ├── websocket/ # Connection & message handling
│ └── models/ # API-facing DTOs (not domain logic)
└── go.mod

Claude should help migrate toward this structure when appropriate, without breaking behavior.

## Architectural rules

### MVC-style responsibilities

- **Controllers**
  - Parse input
  - Validate request shape
  - Call services
  - Return responses
  - No game logic

- **Services**
  - Orchestrate game flow
  - Manage rooms, turns, players
  - Call game logic
  - Enforce rules

- **Game domain**
  - Pure logic
  - Deterministic
  - No HTTP, Gin, or WebSocket awareness
  - Easy to test

### WebSockets

- All WebSocket messages should be:
  - Explicitly typed
  - Versioned if necessary
  - Validated on receipt
- Server must:
  - Verify turn ownership
  - Reject invalid or out-of-order actions
  - Broadcast authoritative state updates

Clients must never be trusted.

## Domain-first development rule

When introducing new backend functionality:

- Always define or update the core domain model first under `internal/game`.
- Game and domain logic must be written without any dependency on Gin, HTTP, or WebSockets.
- Services may then orchestrate domain logic and manage state or lifecycle concerns.
- Controllers must remain thin and only translate HTTP or WebSocket input into service calls.

Claude should not introduce services or controllers without first establishing the underlying domain model.

## Routing & versioning

- Routes are registered in `internal/routes`
- API routes are versioned (e.g. `/api/v1`)
- WebSocket endpoints are versioned alongside HTTP APIs

Example:

- `/api/v1/lobbies`
- `/api/v1/ws/game/:roomCode`

## Error handling

- Always return structured errors
- Wrap errors with context
- Never panic in request paths
- Fail fast on invalid game actions

## State management

- Game state lives server-side only
- Prefer explicit structs over maps
- Avoid shared mutable state without synchronization
- Assume concurrent players

## What NOT to do

- No business logic in Gin handlers
- No game rules in JSON marshaling
- No silent error swallowing
- No client-driven state transitions

## Running & validation

- Run locally via `go run cmd/api/main.go`
- Must compile cleanly
- New logic should be testable without networking
