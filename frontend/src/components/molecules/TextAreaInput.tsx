"use client"; // イベントハンドラを持つためクライアントコンポーネント

import React from 'react';
import { cn } from "@/lib/utils"; // Shadcn/ui スタイルでよく使われるクラス名結合ユーティリティ (なければ作成)
import { Button } from "@/components/atoms/Button"; // 共通 Button をインポート

interface TextAreaInputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
  placeholder?: string;
  isLoading?: boolean;
  buttonText?: string;
  rows?: number;
  className?: string;
}

// --- スタイル定義: Button コンポーネントに内包されたものは不要になる可能性あり ---
const neumorphismBase = "bg-gray-200 rounded-lg transition-all duration-200 ease-in-out";
const neumorphismInsetShadow = "shadow-[inset_5px_5px_10px_#bebebe,inset_-5px_-5px_10px_#ffffff]";
// Button スタイルは Button.tsx に集約
// const neumorphismShadow = "shadow-[5px_5px_10px_#bebebe,_-5px_-5px_10px_#ffffff]";
// const neumorphismButtonHover = "hover:shadow-[3px_3px_6px_#bebebe,_-3px_-3px_6px_#ffffff]";
// const neumorphismButtonActive = "active:shadow-[inset_3px_3px_5px_#bebebe,inset_-3px_-3px_5px_#ffffff]";
// ---

export const TextAreaInput: React.FC<TextAreaInputProps> = ({
  value,
  onChange,
  onSubmit,
  placeholder = "メッセージを入力...",
  isLoading = false,
  buttonText = "送信",
  rows = 3,
  className,
}) => {

  const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Enterのみで送信、Shift+Enterで改行
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault(); // デフォルトの改行動作をキャンセル
      if (!isLoading && value.trim()) {
        onSubmit();
      }
    }
  };

  return (
    <div className={cn("relative p-1", neumorphismBase, neumorphismInsetShadow, className)}>
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        rows={rows}
        disabled={isLoading}
        className={cn(
          "w-full p-3 pr-28 bg-transparent border-none focus:outline-none focus:ring-0 resize-none",
          "placeholder-gray-500 text-gray-700",
          isLoading ? "opacity-50 cursor-not-allowed" : ""
        )}
        aria-label={placeholder}
      />
      <Button
        onClick={() => !isLoading && value.trim() && onSubmit()}
        disabled={!value.trim()} // isLoading は Button 側でハンドリング
        isLoading={isLoading}
        className="absolute bottom-3 right-3" // 位置調整クラスは維持
        size="sm" // サイズを調整 (元の padding/height に合わせる)
        aria-label={buttonText}
      >
        {buttonText} {/* isLoading 時のテキストは Button 側でハンドリング */}
      </Button>
    </div>
  );
}; 