# poke-battles

A real-time, turn-based multiplayer battle game inspired by Pokémon Stadium.

## Testing WebSockets with Postman

### Prerequisites

1. Start the backend server:
   ```bash
   cd backend
   go run cmd/api/main.go
   ```

### Step 1: Create a Lobby (HTTP)

1. Open Postman
2. Create a new **POST** request
3. URL: `http://localhost:8080/api/v1/lobbies`
4. Go to **Body** → **raw** → **JSON**
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

1. Click **New** → **WebSocket**
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

1. Open a new tab → **WebSocket**
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
