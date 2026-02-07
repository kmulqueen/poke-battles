import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createLobby } from "../../../api/lobbies";
import { lobbiesQueryKey } from "./useLobbies";
import type { CreateLobbyRequest } from "../../../types/lobby";

export function useCreateLobby() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateLobbyRequest) => createLobby(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: lobbiesQueryKey });
    },
  });
}
