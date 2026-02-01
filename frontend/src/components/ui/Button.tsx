import type { ButtonHTMLAttributes } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
}

function Button({ type = "button", children, ...props }: ButtonProps) {
  return (
    <button
      type={type}
      className="w-full min-h-11 px-6 py-3 text-base font-semibold text-white bg-primary-500 rounded-xl shadow-md transition-all duration-200 ease-out hover:bg-primary-600 hover:shadow-lg hover:-translate-y-0.5 active:bg-primary-700 active:shadow-md active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary-500 focus-visible:ring-offset-2 disabled:bg-neutral-300 disabled:text-neutral-500 disabled:shadow-none disabled:cursor-not-allowed disabled:hover:translate-y-0"
      {...props}
    >
      {children}
    </button>
  );
}

export default Button;
