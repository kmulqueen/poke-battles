# Backend Repo Context

This skill defines how the backend codebase is organized.

## Language & Frameworks

- Go 1.24
- Gin for HTTP routing
- Gorilla WebSocket for realtime communication

## Directory Structure

- `cmd/api/`
  - Application entrypoint and server wiring
- `internal/controllers/`
  - Thin HTTP handlers
  - No business logic
- `internal/services/`
  - Orchestration and coordination logic
- `internal/game/`
  - Core domain logic
  - Pure, deterministic, testable
- `internal/websocket/`
  - Connection lifecycle
  - Message routing
  - Hub management

## Package Boundaries

- Controllers must not contain business rules
- Services may coordinate multiple packages
- Game domain must not depend on transport (HTTP/WebSocket)
- WebSocket layer adapts messages into domain calls

## Naming Conventions

- Packages: singular, lowercase
- Files: snake_case
- Interfaces: behavior-based, not role-based

## Responsibilities

- Domain logic lives in `game`
- State transitions are explicit
- Side effects are isolated at the edges
