'use client';

import React, { useState } from 'react';
import { Button } from '@/components/atoms/Button';
import { Spinner } from '@/components/atoms/Spinner';
import { TextAreaInput } from '@/components/molecules/TextAreaInput';
import { ResultDisplay } from '@/components/organisms/ResultDisplay';
import { useSummarizeTextMutation } from '@/hooks/api/useSummarizeTextMutation';

const SummarizeTextPage: React.FC = () => {
    const [inputText, setInputText] = useState<string>('');
    const { mutate, data, error, isPending } = useSummarizeTextMutation();

    const handleSummarize = () => {
        if (!inputText.trim()) return; // 空の場合は何もしない
        mutate({ text: inputText });
    };

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold mb-4">自由テキスト要約</h1>
            <p className="mb-4">
                ここにテキストを入力すると、Bedrock が内容を要約します。
            </p>

            <div className="grid grid-cols-1 gap-6">
                {/* Input Area */}
                <div className="space-y-4">
                    <TextAreaInput
                        value={inputText}
                        onChange={setInputText}
                        onSubmit={() => {}}
                        placeholder="要約したいテキストを入力してください..."
                        rows={10}
                        isLoading={isPending}
                    />
                    <Button
                        onClick={handleSummarize}
                        disabled={isPending || !inputText.trim()}
                    >
                        {isPending ? <Spinner size="sm" /> : '要約実行'}
                    </Button>
                </div>

                {/* Result Area */}
                <ResultDisplay
                    title="要約結果"
                    content={data?.summary ?? ''}
                    isLoading={isPending}
                    error={error}
                />
            </div>
        </div>
    );
};

export default SummarizeTextPage; 