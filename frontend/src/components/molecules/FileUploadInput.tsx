"use client";

import React, { useRef, useCallback } from 'react';
import { cn } from "@/lib/utils";
import { Button } from "@/components/atoms/Button";

interface FileUploadInputProps {
  onFileSelect: (file: File) => void;
  acceptedFileTypes?: string;
  label?: string;
  isLoading?: boolean;
  className?: string;
}

// --- スタイル定義 (Button コンポーネントに集約) ---
// const neumorphismBase = ...
// const neumorphismShadow = ...
// const neumorphismButtonHover = ...
// const neumorphismButtonActive = ...
// ---

export const FileUploadInput: React.FC<FileUploadInputProps> = ({
  onFileSelect,
  acceptedFileTypes,
  label = "ファイルを選択",
  isLoading = false,
  className,
}) => {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleButtonClick = () => {
    inputRef.current?.click();
  };

  const handleFileChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (file) {
        onFileSelect(file);
        // 同じファイルを連続して選択できるように input の値をリセット
        event.target.value = "";
      }
    },
    [onFileSelect]
  );

  return (
    <div className={cn("inline-block", className)}>
      <input
        type="file"
        ref={inputRef}
        onChange={handleFileChange}
        accept={acceptedFileTypes}
        className="hidden" // input自体は非表示
        disabled={isLoading}
        aria-hidden="true" // スクリーンリーダーからも隠す
      />
      <Button
        type="button"
        onClick={handleButtonClick}
        isLoading={isLoading}
        aria-label={label}
        // className prop で追加のスタイルを渡せるようにする (必要なら)
        // size="default" // デフォルトサイズを使用
      >
        {label} {/* isLoading 時の表示は Button 側でハンドリング */}
      </Button>
    </div>
  );
}; 