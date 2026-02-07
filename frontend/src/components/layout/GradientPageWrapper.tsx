import type { PropsWithChildren } from "react";

interface GradientPageWrapperProps extends PropsWithChildren {
  className?: string;
}

function GradientPageWrapper({ children, className = "" }: GradientPageWrapperProps) {
  return (
    <div
      className={`min-h-dvh bg-gradient-to-br from-primary-500 via-primary-600 to-secondary-700 flex flex-col ${className}`}
    >
      {children}
    </div>
  );
}

export default GradientPageWrapper;
