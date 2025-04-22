'use client';

import React from 'react';
import { cn } from "@/lib/utils";
import { ApiError } from '@/lib/apiClient';
import type { DocumentSnippet } from '@/types/api';
import { Spinner } from '@/components/atoms/Spinner';

interface RecommendationDisplayProps {
    title?: string;
    recommendations: DocumentSnippet[] | undefined; // Accept undefined for initial state
    isLoading?: boolean;
    error?: ApiError | Error | null;
    className?: string;
}

// Reusing neumorphism styles (consider consolidating styles later)
const neumorphismBase = "bg-gray-200 rounded-lg transition-all duration-200 ease-in-out";
const neumorphismInsetShadow = "shadow-[inset_5px_5px_10px_#bebebe,inset_-5px_-5px_10px_#ffffff]";
const neumorphismSourceShadow = "shadow-[3px_3px_6px_#bebebe,_-3px_-3px_6px_#ffffff]";

export const RecommendationDisplay: React.FC<RecommendationDisplayProps> = ({
    title = "関連文書",
    recommendations,
    isLoading = false,
    error = null,
    className,
}) => {

    const renderContent = () => {
        if (isLoading) {
            return (
                <div className="flex justify-center items-center h-20 text-gray-500">
                    <Spinner size="default" />
                    <span className="ml-2">関連文書を検索中...</span>
                </div>
            );
        }

        if (error) {
            const errorMessage = error instanceof ApiError ? `${error.message} (Status: ${error.status})` : error.message;
            return (
                <div className="text-red-600 p-4 border border-red-300 rounded-md bg-red-50">
                    <p className="font-semibold">関連文書の取得エラー</p>
                    <p className="text-sm">{errorMessage}</p>
                </div>
            );
        }

        if (!recommendations || recommendations.length === 0) {
            return <p className="text-gray-500 text-sm p-4">関連する文書は見つかりませんでした。</p>;
        }

        return (
            <ul className="space-y-3 p-1">
                {recommendations.map((rec, index) => (
                    <li key={rec.document_id + index} className={`p-3 rounded-md ${neumorphismBase} ${neumorphismSourceShadow}`}>
                        {rec.title && <p className="text-xs font-medium text-gray-700 mb-1">{rec.title}</p>}
                        <p className="text-xs text-gray-600 italic">&quot;{rec.snippet}&quot;</p>
                        {rec.score && <p className="text-xs text-gray-500 mt-1">Score: {rec.score.toFixed(4)}</p>}
                        {/* Add link to document detail page if available */}
                    </li>
                ))}
            </ul>
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
            aria-live="polite"
        >
            {title && <h3 className="text-lg font-semibold text-gray-700 mb-3">{title}</h3>}
            {renderContent()}
        </div>
    );
}; 