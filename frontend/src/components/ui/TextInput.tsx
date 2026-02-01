import type { InputHTMLAttributes } from "react";

interface TextInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "id"> {
  id: string;
  label: string;
  onChange: (value: string) => void;
  wrapperClassName?: string;
  error?: string;
  helperText?: string;
}

function TextInput({ id, label, onChange, wrapperClassName, error, helperText, ...props }: TextInputProps) {
  const errorId = error ? `${id}-error` : undefined;
  const helperTextId = helperText ? `${id}-helper` : undefined;
  const describedBy = [errorId, helperTextId].filter(Boolean).join(" ") || undefined;

  return (
    <div className={wrapperClassName ? wrapperClassName : "flex flex-col gap-2"}>
      <label htmlFor={id} className="text-sm font-medium text-neutral-700">
        {label}
      </label>
      <input
        type="text"
        id={id}
        name={id}
        className={`w-full min-h-11 px-4 py-3 text-base text-neutral-900 placeholder:text-neutral-400 bg-surface rounded-xl shadow-sm transition-colors duration-200 ${
          error
            ? "border border-red-500 hover:border-red-600 focus-visible:border-red-500 focus-visible:ring-2 focus-visible:ring-red-500/50 focus-visible:outline-none"
            : "border border-neutral-300 hover:border-neutral-400 focus-visible:border-secondary-500 focus-visible:ring-2 focus-visible:ring-secondary-500/50 focus-visible:outline-none"
        }`}
        onChange={(e) => onChange(e.target.value)}
        aria-describedby={describedBy}
        aria-invalid={error ? true : undefined}
        {...props}
      />
      {helperText && (
        <p id={helperTextId} className="text-sm text-neutral-600">
          {helperText}
        </p>
      )}
      {error && (
        <p id={errorId} className="text-sm text-red-600" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}

export default TextInput;
