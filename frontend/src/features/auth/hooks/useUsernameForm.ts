import { useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "@tanstack/react-router";

export function useUsernameForm() {
  const [username, setUsername] = useState("");
  const navigate = useNavigate();

  const isValid = username.trim().length >= 3;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!isValid) return;
    // Future: integrate with auth API or WebSocket connection
    // This will eventually authenticate the user and store credentials
    // before navigating to the lobby
    navigate({ to: "/lobby" });
  };

  return {
    username,
    setUsername,
    handleSubmit,
    isValid,
  };
}
