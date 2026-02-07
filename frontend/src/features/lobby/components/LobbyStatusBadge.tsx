import type { LobbyState } from "../../../types/lobby";

interface LobbyStatusBadgeProps {
  status: LobbyState;
}

const statusConfig = {
  waiting: {
    bg: "bg-accent-400/20",
    text: "text-accent-600",
    border: "border-accent-500",
    label: "Waiting",
  },
  ready: {
    bg: "bg-green-100",
    text: "text-green-700",
    border: "border-green-400",
    label: "Ready",
  },
  active: {
    bg: "bg-secondary-100",
    text: "text-secondary-700",
    border: "border-secondary-400",
    label: "In Game",
  },
} as const;

function LobbyStatusBadge({ status }: LobbyStatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${config.bg} ${config.text} ${config.border}`}
    >
      <span className="sr-only">Lobby status: </span>
      {config.label}
    </span>
  );
}

export default LobbyStatusBadge;
