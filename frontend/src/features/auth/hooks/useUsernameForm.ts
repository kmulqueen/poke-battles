import { useState } from "react";
import type { FormEvent } from "react";

export function useUsernameForm() {
  const [username, setUsername] = useState("");

  const isValid = username.trim().length >= 3;

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (!isValid) return;
    // Future: integrate with auth API or WebSocket connection
    alert(`Sign up functionality needed!`);
  };

  return {
    username,
    setUsername,
    handleSubmit,
    isValid,
  };
}
