import Button from "../../../components/ui/Button";

interface LobbiesErrorStateProps {
  error: string;
  onRetry: () => void;
}

function LobbiesErrorState({ error, onRetry }: LobbiesErrorStateProps) {
  return (
    <div className="bg-surface rounded-2xl shadow-xl py-16 px-4" role="alert">
      <div className="flex flex-col items-center justify-center text-center">
        <div className="w-16 h-16 mb-4 text-red-500">
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
              d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
            />
          </svg>
        </div>
        <h2 className="text-xl font-semibold text-neutral-700 mb-2">Failed to load lobbies</h2>
        <p className="text-neutral-600 mb-6 max-w-md">{error}</p>
        <div className="w-full max-w-xs">
          <Button onClick={onRetry}>Try Again</Button>
        </div>
      </div>
    </div>
  );
}

export default LobbiesErrorState;
