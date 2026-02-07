import Button from "../../../components/ui/Button";

interface LobbiesEmptyStateProps {
  isFiltered: boolean;
  onCreateLobby: () => void;
}

function LobbiesEmptyState({ isFiltered, onCreateLobby }: LobbiesEmptyStateProps) {
  return (
    <div className="bg-surface rounded-2xl shadow-xl py-16 px-4" role="status">
      <div className="flex flex-col items-center justify-center text-center">
        <div className="w-16 h-16 mb-4 text-neutral-400">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            strokeWidth={1.5}
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z"
            />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-neutral-700 mb-2">No lobbies found</h2>
        <p className="text-neutral-600 mb-6 max-w-md">
          {isFiltered
            ? "Try adjusting your filters to see more lobbies."
            : "There are no active lobbies right now. Create one to get started!"}
        </p>
        {!isFiltered && (
          <div className="w-full max-w-xs">
            <Button onClick={onCreateLobby}>Create Lobby</Button>
          </div>
        )}
      </div>
    </div>
  );
}

export default LobbiesEmptyState;
