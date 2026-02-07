import { useQuery } from "@tanstack/react-query";
import { fetchLobbies } from "../../../api/lobbies";

export const lobbiesQueryKey = ["lobbies"] as const;

export function useLobbies() {
  return useQuery({
    queryKey: lobbiesQueryKey,
    queryFn: fetchLobbies,
    staleTime: 10_000, // 10 seconds
    refetchInterval: 30_000, // Refresh every 30 seconds
  });
}
