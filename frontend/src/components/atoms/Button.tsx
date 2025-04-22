"use client";

import React from 'react';
import { Slot } from '@radix-ui/react-slot'; // Optional: Allows wrapping other components
import { cva, type VariantProps } from 'class-variance-authority'; // For handling variants

import { cn } from "@/lib/utils";

// --- Neumorphism Styles (共通化推奨) ---
const neumorphismBase = "bg-gray-200 rounded-lg transition-all duration-200 ease-in-out";
const neumorphismShadow = "shadow-[5px_5px_10px_#bebebe,_-5px_-5px_10px_#ffffff]";
const neumorphismHover = "hover:shadow-[3px_3px_6px_#bebebe,_-3px_-3px_6px_#ffffff]";
const neumorphismActive = "active:shadow-[inset_3px_3px_5px_#bebebe,inset_-3px_-3px_5px_#ffffff]";
const neumorphismDisabled = "shadow-[inset_2px_2px_4px_#bebebe,inset_-2px_-2px_4px_#ffffff] opacity-50 cursor-not-allowed";
// ---

const buttonVariants = cva(
  // Base styles for all buttons
  [
    "inline-flex items-center justify-center whitespace-nowrap",
    "font-semibold",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2 focus-visible:ring-offset-gray-200",
    neumorphismBase,
    neumorphismShadow,
  ],
  {
    variants: {
      variant: {
        default: [
          "text-gray-700",
          neumorphismHover,
          neumorphismActive,
        ],
        // primary: "bg-blue-500 text-white ...", // 将来の拡張用
        // destructive: "bg-red-500 text-white ...", // 将来の拡張用
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        icon: "h-10 w-10", // アイコンボタン用
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean; // To use Slot
  isLoading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant,
      size,
      asChild = false,
      isLoading = false,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    const Comp = asChild ? Slot : "button";
    const isDisabled = disabled || isLoading;

    return (
      <Comp
        className={cn(
          buttonVariants({ variant, size, className }),
          isDisabled ? neumorphismDisabled : ""
        )}
        ref={ref}
        disabled={isDisabled}
        aria-disabled={isDisabled}
        {...props}
      >
        {isLoading ? (
          // TODO: より良いローディングスピナーを実装
          <span className="animate-pulse">処理中...</span>
        ) : (
          children
        )}
      </Comp>
    );
  }
);
Button.displayName = "Button";

export { Button, buttonVariants }; 