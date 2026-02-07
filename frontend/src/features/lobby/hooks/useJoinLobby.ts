import { useMutation, useQueryClient } from "@tanstack/react-query";
import { joinLobby } from "../../../api/lobbies";
import { lobbiesQueryKey } from "./useLobbies";
import type { JoinLobbyRequest } from "../../../types/lobby";

interface JoinLobbyParams {
  code: string;
  request: JoinLobbyRequest;
}

export function useJoinLobby() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ code, request }: JoinLobbyParams) => joinLobby(code, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: lobbiesQueryKey });
    },
  });
}
