import { get, post } from "./http";
import type { Lobby, CreateLobbyRequest, JoinLobbyRequest } from "../types/lobby";

export async function fetchLobbies(): Promise<Lobby[]> {
  return get<Lobby[]>("/lobbies");
}

export async function fetchLobbyByCode(code: string): Promise<Lobby> {
  return get<Lobby>(`/lobbies/${code}`);
}

export async function createLobby(request: CreateLobbyRequest): Promise<Lobby> {
  return post<Lobby, CreateLobbyRequest>("/lobbies", request);
}

export async function joinLobby(
  code: string,
  request: JoinLobbyRequest
): Promise<Lobby> {
  return post<Lobby, JoinLobbyRequest>(`/lobbies/${code}/join`, request);
}
