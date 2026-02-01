# Frontend (React / TypeScript / Vite) Instructions

This frontend is a real-time multiplayer game client that renders server-authoritative state and sends player intent.

## Core technologies

- React + TypeScript
- Vite
- Redux Toolkit for client state
- React Query for server state
- Tailwind CSS for styling
- WebSockets for real-time updates

## Core principles

- The frontend never calculates battle outcomes
- Redux stores client/UI state
- React Query manages server state
- WebSocket messages are treated as events, not commands

## State management rules

### Redux Toolkit

Use Redux for:

- UI state (modals, menus, selections)
- Player input (selected move, target)
- Connection status

Do NOT use Redux for:

- Battle resolution
- Turn logic
- Damage calculations

### React Query

Use React Query for:

- Lobbies
- Room metadata
- Initial game state hydration
- Any HTTP-based server data

Server state should not be copied into Redux unless necessary.

## WebSockets

- WebSocket messages update state based on server events
- Never optimistically apply game outcomes
- Always reconcile with server state

## Component rules

- Components should be presentational where possible
- Hooks encapsulate logic
- Avoid large monolithic components
- Prefer composition over prop drilling

## Styling

- Tailwind CSS only
- No inline styles unless necessary
- Prefer semantic class grouping
- Avoid custom CSS unless Tailwind cannot express it

## File organization (future-facing)

frontend/src/
├── features/
│ ├── lobby/
│ ├── battle/
│ └── matchmaking/
├── components/
├── hooks/
├── api/
│ ├── http.ts
│ └── ws.ts
├── routes/
├── store/
└── types/

Claude should help migrate toward this structure incrementally.

## What to avoid

- Duplicating backend game rules
- Hardcoding timing assumptions
- Deeply nested state
- Side effects inside render paths

## Validation

- Must type-check
- No unused imports
- ESLint rules should pass
