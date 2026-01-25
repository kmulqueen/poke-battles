# Claude Project Instructions

This repository contains a full-stack web application for a real-time, turn-based multiplayer game inspired by Pokémon Stadium.

## High-level architecture

- Go (Gin) backend
- React + TypeScript + Vite frontend
- Docker Compose for local orchestration
- WebSockets for real-time gameplay
- Server is authoritative for all game logic

## Core principles

- The backend is the single source of truth for game state.
- Clients never calculate battle outcomes.
- All real-time communication must be deterministic and replay-safe.
- Prefer clarity and explicitness over clever abstractions.

## General rules for changes

- Do not introduce new frameworks or libraries without justification.
- Keep backend and frontend concerns strictly separated.
- Avoid speculative refactors — only change what is necessary.
- Favor small, composable functions over large multi-purpose ones.
- If unsure about intent, preserve existing behavior.

## Project layout expectations

- Go backend code lives under `backend/`
- Frontend code lives under `frontend/`
- Backend follows a `cmd/` + `internal/` layout
- Frontend follows feature-based organization once it grows

## Validation expectations

When modifying code:

- Backend changes should compile (`go build ./...`)
- Frontend changes should type-check and build
- Avoid leaving unused code, imports, or dead files

## Things to avoid

- Duplicating business logic between client and server
- Embedding game rules in HTTP or WebSocket handlers
- Making frontend assumptions about server timing or ordering
