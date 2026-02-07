import { useState, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import PageHeader from "../components/layout/PageHeader";
import GradientPageWrapper from "../components/layout/GradientPageWrapper";
import Button from "../components/ui/Button";
import TextInput from "../components/ui/TextInput";
import Toggle from "../components/ui/Toggle";
import { useLobbies } from "../features/lobby/hooks/useLobbies";
import LobbyToolbar from "../features/lobby/components/LobbyToolbar";
import LobbyCard from "../features/lobby/components/LobbyCard";
import LobbiesLoadingState from "../features/lobby/components/LobbiesLoadingState";
import LobbiesErrorState from "../features/lobby/components/LobbiesErrorState";
import LobbiesEmptyState from "../features/lobby/components/LobbiesEmptyState";

function LobbiesPage() {
  const navigate = useNavigate();
  const { data: lobbies, isLoading, isError, error, refetch } = useLobbies();

  const [searchQuery, setSearchQuery] = useState("");
  const [showJoinableOnly, setShowJoinableOnly] = useState(false);

  const filteredLobbies = useMemo(() => {
    if (!lobbies) return [];

    let filtered = lobbies;

    // Filter by search query (lobby code)
    if (searchQuery.trim()) {
      filtered = filtered.filter((lobby) =>
        lobby.code.toLowerCase().includes(searchQuery.toLowerCase()),
      );
    }

    // Filter by joinable status
    if (showJoinableOnly) {
      filtered = filtered.filter(
        (lobby) =>
          lobby.state === "waiting" && lobby.players.length < lobby.max_players,
      );
    }

    return filtered;
  }, [lobbies, searchQuery, showJoinableOnly]);

  const isFiltered = searchQuery.trim() !== "" || showJoinableOnly;

  const handleCreateLobby = () => {
    navigate({ to: "/lobby/create" });
  };

  const handleJoinLobby = (code: string) => {
    navigate({ to: "/lobby/$code/join", params: { code } });
  };

  return (
    <GradientPageWrapper>
      <PageHeader title="Game Lobbies" />
      <main className="flex-1 w-full max-w-4xl mx-auto px-4 sm:px-6 md:px-8 pb-8">
        {/* Toolbar */}
        <LobbyToolbar>
          <div className="flex-1 w-full">
            <TextInput
              id="search-lobby"
              label="Search by code"
              value={searchQuery}
              onChange={setSearchQuery}
              placeholder="Enter lobby code..."
            />
          </div>
          <div className="flex items-center sm:items-end gap-4 w-full sm:w-auto">
            <Toggle
              id="joinable-only"
              label="Joinable only"
              checked={showJoinableOnly}
              onChange={setShowJoinableOnly}
            />
            <div className="sm:w-40 w-full">
              <Button onClick={handleCreateLobby}>Create Lobby</Button>
            </div>
          </div>
        </LobbyToolbar>

        {/* Content */}
        <section aria-label="Available lobbies">
          {/* Live region for filter result announcements */}
          <div className="sr-only" aria-live="polite" aria-atomic="true">
            {!isLoading &&
              !isError &&
              `${filteredLobbies.length} lobbies found`}
          </div>

          {isLoading && <LobbiesLoadingState />}

          {isError && (
            <LobbiesErrorState
              error={error?.message || "An unexpected error occurred"}
              onRetry={() => refetch()}
            />
          )}

          {!isLoading && !isError && filteredLobbies.length === 0 && (
            <LobbiesEmptyState
              isFiltered={isFiltered}
              onCreateLobby={handleCreateLobby}
            />
          )}

          {!isLoading && !isError && filteredLobbies.length > 0 && (
            <ul className="space-y-4">
              {filteredLobbies.map((lobby) => (
                <li key={lobby.code}>
                  <LobbyCard lobby={lobby} onJoin={handleJoinLobby} />
                </li>
              ))}
            </ul>
          )}
        </section>
      </main>
    </GradientPageWrapper>
  );
}

export default LobbiesPage;
