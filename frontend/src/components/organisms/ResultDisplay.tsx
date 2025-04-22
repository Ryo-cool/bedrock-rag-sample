"use client";

import React from 'react';
import { cn } from "@/lib/utils";
import { ApiError } from '@/lib/apiClient';
import type { DocumentSnippet } from '@/types/api';
import { Spinner } from '@/components/atoms/Spinner';

interface ResultDisplayProps {
  title?: string;
  content: string | React.ReactNode;
  isLoading?: boolean;
  error?: ApiError | Error | null; // ApiError または標準 Error を受け入れる
  sources?: DocumentSnippet[];
  className?: string;
}

// ニューモフィズム スタイル定義
const neumorphismBase = "bg-gray-200 rounded-lg transition-all duration-200 ease-in-out";
const neumorphismInsetShadow = "shadow-[inset_5px_5px_10px_#bebebe,inset_-5px_-5px_10px_#ffffff]";
const neumorphismSourceShadow = "shadow-[3px_3px_6px_#bebebe,_-3px_-3px_6px_#ffffff]";

export const ResultDisplay: React.FC<ResultDisplayProps> = ({
  title,
  content,
  isLoading = false,
  error = null,
  sources,
  className,
}) => {
  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex justify-center items-center h-20 text-gray-500">
          <Spinner size="default" />
        </div>
      );
    }

    if (error) {
      // ApiError かどうかで表示を分けることも可能
      const errorMessage = error instanceof ApiError ? `${error.message} (Status: ${error.status})` : error.message;
      return (
        <div className="text-red-600 p-4 border border-red-300 rounded-md bg-red-50">
          <p className="font-semibold">エラーが発生しました</p>
          <p className="text-sm">{errorMessage}</p>
          {/* ApiError の場合、詳細情報を表示することも検討 error.info */}
        </div>
      );
    }

    if (content) {
      return (
        <div className="prose prose-sm max-w-none text-gray-800">
          {typeof content === 'string' ? (
            <p className="whitespace-pre-wrap">{content}</p>
          ) : (
            content
          )}
        </div>
      );
    }

    return null; // 何も表示しない場合
  };

  const renderSources = () => {
    if (!sources || sources.length === 0 || isLoading || error) {
      return null;
    }

    return (
      <div className="mt-6 pt-4 border-t border-gray-300">
        <h4 className="text-sm font-semibold text-gray-600 mb-2">参照元:</h4>
        <ul className="space-y-2">
          {sources.map((source, index) => (
            <li key={source.document_id + index} className={`p-3 rounded-md ${neumorphismBase} ${neumorphismSourceShadow}`}>
              {source.title && <p className="text-xs font-medium text-gray-700 mb-1">{source.title}</p>}
              <p className="text-xs text-gray-600 italic">&quot;{source.snippet}&quot;</p>
              {/* source_url があればリンクにするなど */}
            </li>
          ))}
        </ul>
      </div>
    );
  };

  return (
    <div
      className={cn(
        "p-4",
        neumorphismBase,
        neumorphismInsetShadow,
        className
      )}
      aria-live="polite" // 非同期コンテンツ更新を通知
    >
      {title && <h3 className="text-lg font-semibold text-gray-700 mb-3">{title}</h3>}
      {renderContent()}
      {renderSources()}
    </div>
  );
}; 