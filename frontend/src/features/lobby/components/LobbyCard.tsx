import type { Lobby } from "../../../types/lobby";
import Button from "../../../components/ui/Button";
import LobbyStatusBadge from "./LobbyStatusBadge";
import pokeBallsImage from "../../../assets/poke-balls.png";

interface LobbyCardProps {
  lobby: Lobby;
  onJoin: (code: string) => void;
}

function LobbyCard({ lobby, onJoin }: LobbyCardProps) {
  const isJoinable =
    lobby.state === "waiting" && lobby.players.length < lobby.max_players;
  const disabledReasonId = `lobby-${lobby.code}-disabled-reason`;

  const getDisabledReason = () => {
    if (lobby.state !== "waiting") {
      return "Game in progress";
    }
    if (lobby.players.length >= lobby.max_players) {
      return "Lobby is full";
    }
    return "";
  };

  const disabledReason = !isJoinable ? getDisabledReason() : "";

  return (
    <article
      className="relative overflow-hidden bg-surface rounded-2xl shadow-lg p-4 sm:p-6 hover:shadow-xl hover:-translate-y-0.5 transform-gpu transition-all duration-200"
      aria-labelledby={`lobby-code-${lobby.code}`}
    >
      <div
        className="absolute inset-0 opacity-10 sm:opacity-15 pointer-events-none"
        aria-hidden="true"
        style={{
          backgroundImage: `url(${pokeBallsImage})`,
          backgroundRepeat: "repeat",
          backgroundPosition: "right center",
          backgroundSize: "clamp(120px, 50%, 750px) auto",
        }}
      />
      <div className="relative z-10 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex-1 space-y-2">
          <div className="flex items-center gap-2">
            <LobbyStatusBadge status={lobby.state} />
          </div>
          <div>
            <p className="text-sm text-neutral-600">Lobby Code</p>
            <h3
              id={`lobby-code-${lobby.code}`}
              className="text-2xl font-bold text-neutral-900 tracking-wide"
            >
              {lobby.code}
            </h3>
          </div>
          <p className="text-sm text-neutral-600">
            Players: {lobby.players.length}/{lobby.max_players}
          </p>
          {disabledReason && (
            <p id={disabledReasonId} className="sr-only">
              {disabledReason}
            </p>
          )}
        </div>
        <div className="sm:w-40">
          <Button
            onClick={() => onJoin(lobby.code)}
            disabled={!isJoinable}
            aria-describedby={!isJoinable ? disabledReasonId : undefined}
          >
            Join
          </Button>
        </div>
      </div>
    </article>
  );
}

export default LobbyCard;
