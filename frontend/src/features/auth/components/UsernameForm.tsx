import type { FormEvent } from "react";
import Button from "../../../components/ui/Button";
import TextInput from "../../../components/ui/TextInput";

interface UsernameFormProps {
  username: string;
  onUsernameChange: (value: string) => void;
  onSubmit: (e: FormEvent) => void;
  disabled?: boolean;
}

function UsernameForm({ username, onUsernameChange, onSubmit, disabled }: UsernameFormProps) {
  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-5 sm:gap-6">
      <TextInput
        id="input-username"
        label="Username"
        value={username}
        placeholder="Enter a username..."
        onChange={onUsernameChange}
        helperText="Username must be at least 3 characters"
        autoComplete="username"
        required
      />
      <Button type="submit" disabled={disabled}>Enter</Button>
    </form>
  );
}

export default UsernameForm;
