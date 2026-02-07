import type { ButtonHTMLAttributes } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  variant?: "primary" | "secondary";
}

function Button({ type = "button", variant = "primary", children, ...props }: ButtonProps) {
  const baseClasses = "w-full min-h-11 px-6 py-3 text-base font-semibold rounded-xl shadow-md transition-all duration-200 ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary-500 focus-visible:ring-offset-2 disabled:shadow-none disabled:cursor-not-allowed disabled:hover:translate-y-0";

  const variantClasses = {
    primary: "text-white bg-primary-500 hover:bg-primary-600 hover:shadow-lg hover:-translate-y-0.5 active:bg-primary-700 active:shadow-md active:translate-y-0 disabled:bg-neutral-300 disabled:text-neutral-500",
    secondary: "text-primary-700 bg-white border-2 border-primary-500 hover:bg-primary-50 hover:shadow-lg hover:-translate-y-0.5 active:bg-primary-100 active:shadow-md active:translate-y-0 disabled:bg-neutral-100 disabled:text-neutral-400 disabled:border-neutral-300"
  };

  return (
    <button
      type={type}
      className={`${baseClasses} ${variantClasses[variant]}`}
      {...props}
    >
      {children}
    </button>
  );
}

export default Button;
