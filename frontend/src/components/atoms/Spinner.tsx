import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from "@/lib/utils";

const spinnerVariants = cva(
  "animate-spin rounded-full border-solid border-current border-t-transparent",
  {
    variants: {
      size: {
        default: "h-6 w-6 border-4",
        sm: "h-4 w-4 border-2",
        lg: "h-10 w-10 border-4",
      },
    },
    defaultVariants: {
      size: "default",
    },
  }
);

interface SpinnerProps extends VariantProps<typeof spinnerVariants> {
  className?: string;
}

export const Spinner: React.FC<SpinnerProps> = ({
  className,
  size,
}) => {
  return (
    <div
      className={cn(spinnerVariants({ size, className }))}
      role="status"
      aria-live="polite"
      aria-label="読み込み中"
    />
  );
}; 