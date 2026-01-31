# poke-battles

A real-time, turn-based multiplayer battle game inspired by Pokemon Stadium.

## Tech Stack

- **Backend:** Go 1.24, Gin, Gorilla WebSocket
- **Frontend:** React 19, TypeScript 5.9, Vite
- **Infrastructure:** Docker Compose, Nginx, GitHub Actions CI

## Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose (optional, for containerized development)

## Quick Start

### Local Development (Recommended)

Run backend and frontend in separate terminals:

```bash
# Terminal 1 - Backend (http://localhost:8080)
make dev-backend

# Terminal 2 - Frontend (http://localhost:5173)
make dev-frontend
```

The frontend dev server proxies `/api` requests to the backend automatically.

### Docker Development

```bash
make docker-build   # Build images
make docker-up      # Start containers (backend: 8080, frontend: 3000)
make docker-down    # Stop containers
```

## Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build both backend and frontend |
| `make build-backend` | Compile Go binaries |
| `make build-frontend` | Build React/Vite production bundle |
| `make test` | Run backend tests |
| `make lint` | Lint frontend with ESLint |
| `make dev-backend` | Run backend dev server |
| `make dev-frontend` | Run frontend dev server |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make docker-build` | Build Docker images |

## Project Structure

```
poke-battles/
├── backend/
│   ├── cmd/api/             # Application entrypoint
│   └── internal/
│       ├── controllers/     # HTTP handlers (thin layer)
│       ├── game/            # Core domain logic (pure, testable)
│       ├── services/        # Business orchestration
│       ├── websocket/       # WebSocket hub & connections
│       ├── middleware/      # CORS, etc.
│       └── routes/          # Route registration
├── frontend/
│   └── src/                 # React application
├── .github/workflows/       # CI/CD pipelines
├── Makefile                 # Development commands
└── docker-compose.yml       # Container orchestration
```

## API Endpoints

Base path: `/api/v1`

### HTTP

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/health` | Health check |
| POST | `/lobbies` | Create a new lobby |
| GET | `/lobbies/:code` | Get lobby state |
| POST | `/lobbies/:code/join` | Join an existing lobby |
| POST | `/lobbies/:code/leave` | Leave a lobby |
| POST | `/lobbies/:code/start` | Start game (host only) |

### WebSocket

| Endpoint | Description |
|----------|-------------|
| `/ws/game/:code` | Connect to a game room |

## Testing

```bash
# Run all backend tests
make test

# Run with verbose output
cd backend && go test ./... -v

# Lint frontend
make lint
```

---

## WebSocket Testing with Postman

### Step 1: Create a Lobby (HTTP)

1. Open Postman
2. Create a new **POST** request
3. URL: `http://localhost:8080/api/v1/lobbies`
4. Go to **Body** > **raw** > **JSON**
5. Enter:
   ```json
   {
     "player_id": "player1",
     "username": "Alice"
   }
   ```
6. Click **Send**
7. Copy the `code` from the response (e.g., `"ABC123"`)

### Step 2: Add a Second Player (HTTP)

1. Create another **POST** request
2. URL: `http://localhost:8080/api/v1/lobbies/{CODE}/join` (replace `{CODE}` with your lobby code)
3. Body:
   ```json
   {
     "player_id": "player2",
     "username": "Bob"
   }
   ```
4. Click **Send**

### Step 3: Connect Player 1 via WebSocket

1. Click **New** > **WebSocket**
2. URL: `ws://localhost:8080/api/v1/ws/game/{CODE}` (replace `{CODE}`)
3. Click **Connect**
4. In the message field, send:
   ```json
   {
     "type": "authenticate",
     "version": 1,
     "timestamp": 1706000000000,
     "correlation_id": "auth-player1",
     "payload": {
       "player_id": "player1",
       "session_token": "dummy",
       "lobby_code": "{CODE}"
     }
   }
   ```
5. You should receive `authenticated` and `lobby_updated` messages

### Step 4: Connect Player 2 via WebSocket

1. Open a new tab > **WebSocket**
2. URL: `ws://localhost:8080/api/v1/ws/game/{CODE}`
3. Click **Connect**
4. Send:
   ```json
   {
     "type": "authenticate",
     "version": 1,
     "timestamp": 1706000000000,
     "correlation_id": "auth-player2",
     "payload": {
       "player_id": "player2",
       "session_token": "dummy",
       "lobby_code": "{CODE}"
     }
   }
   ```

### Step 5: Test Two-Way Communication

In Player 1's tab, send:
```json
{
  "type": "set_ready",
  "version": 1,
  "timestamp": 1706000000000,
  "correlation_id": "ready-1",
  "payload": {
    "ready": true
  }
}
```

**Both tabs** should receive a `lobby_updated` message with `"event": "player_ready_changed"`.

### Other Test Messages

**Heartbeat:**
```json
{
  "type": "heartbeat",
  "version": 1,
  "timestamp": 1706000000000,
  "payload": {}
}
```

**Request Lobby State:**
```json
{
  "type": "request_lobby_state",
  "version": 1,
  "timestamp": 1706000000000,
  "payload": {}
}
```

### Message Protocol

All messages require:
- `type` - message type string
- `version` - must be `1`
- `timestamp` - Unix milliseconds
- `payload` - object specific to message type
- `correlation_id` - (optional) for request/response tracking
