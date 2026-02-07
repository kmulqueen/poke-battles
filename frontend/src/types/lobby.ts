export type LobbyState = "waiting" | "ready" | "active";

export interface Player {
  id: string;
  username: string;
}

export interface Lobby {
  code: string;
  state: LobbyState;
  players: Player[];
  host_id: string;
  max_players: number;
}

export interface CreateLobbyRequest {
  player_id: string;
  username: string;
}

export interface JoinLobbyRequest {
  player_id: string;
  username: string;
}
