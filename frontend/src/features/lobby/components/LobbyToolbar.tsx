import { type PropsWithChildren } from "react";

function LobbyToolbar({ children }: PropsWithChildren) {
  return (
    <search className="bg-surface rounded-2xl shadow-xl p-4 sm:p-6 mb-6" aria-label="Lobby filters">
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-end">
        {children}
      </div>
    </search>
  );
}

export default LobbyToolbar;
