# Backend Architecture

This skill defines backend architectural conventions.

Key rules:

- Controllers are thin
- Domain logic lives in services or game
- Game package is pure
- WebSocket layer has no game logic
- WebSocket handlers delegate logic
