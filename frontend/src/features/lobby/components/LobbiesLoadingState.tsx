function LobbiesLoadingState() {
  return (
    <div
      className="bg-surface rounded-2xl shadow-xl py-16 px-4"
      role="status"
      aria-busy="true"
      aria-live="polite"
    >
      <div className="flex flex-col items-center justify-center">
        <div className="w-12 h-12 mb-4">
          <svg
            className="animate-spin text-secondary-500"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>
        <p className="text-neutral-600 font-medium">Loading lobbies...</p>
      </div>
    </div>
  );
}

export default LobbiesLoadingState;
