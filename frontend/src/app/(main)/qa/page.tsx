'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/atoms/Button';
import { Spinner } from '@/components/atoms/Spinner';
import { TextAreaInput } from '@/components/molecules/TextAreaInput';
import { ResultDisplay } from '@/components/organisms/ResultDisplay';
import { useQaMutation } from '@/hooks/api/useQaMutation';
import { cn } from '@/lib/utils'; // For conditional styling
import type { QaResponse, Message, DocumentSnippet } from '@/types/api'; // Import necessary types

// Define source type expected by ResultDisplay
// type ResultDisplaySource = DocumentSnippet;

// Keep ChatMessage source type aligned with API response for now
interface ChatMessage {
    sender: 'user' | 'ai';
    content: string | React.ReactNode;
    sources?: QaResponse['sources']; // Use API response type here
}

const QAPage: React.FC = () => {
    const [inputText, setInputText] = useState<string>('');
    const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
    const { mutate, data, error, isPending, reset } = useQaMutation();
    const chatEndRef = useRef<HTMLDivElement>(null); // Ref to scroll to the bottom

    // Scroll to bottom when chat history updates
    useEffect(() => {
        chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [chatHistory]);

    // Add AI response to chat history when data arrives
    useEffect(() => {
        if (data) {
            setChatHistory((prev) => [
                ...prev,
                { sender: 'ai', content: data.answer, sources: data.sources }, // Keep API source format in state
            ]);
            reset(); // Reset mutation state after processing
        }
    }, [data, reset]);

    // Add error message to chat history
    useEffect(() => {
        if (error) {
            setChatHistory((prev) => [
                ...prev,
                { sender: 'ai', content: <p className="text-red-500">エラー: {error.message}</p> },
            ]);
            reset(); // Reset mutation state after processing
        }
    }, [error, reset]);


    const handleSendMessage = () => {
        if (!inputText.trim() || isPending) return;

        const userMessage: ChatMessage = { sender: 'user', content: inputText };
        // Create message history for API request (currently just the latest message)
        const apiMessages: Message[] = [{ role: 'user', content: inputText }];

        setChatHistory((prev) => [...prev, userMessage]);
        // mutate({ query: inputText }); // Incorrect request format
        mutate({ messages: apiMessages }); // Corrected: Pass messages array
        setInputText(''); // Clear input field
    };

    // Helper function to convert API sources to DocumentSnippet for ResultDisplay
    const mapSourcesToDocumentSnippets = (apiSources: QaResponse['sources']): DocumentSnippet[] | undefined => {
        return apiSources?.map((source, index) => ({
            document_id: source.source_url || `source-${index}`, // Generate dummy ID if URL is missing
            snippet: source.snippet,
            // title and score are optional in DocumentSnippet
            // source_url could potentially be used for title if needed
        }));
    };

    // Styles for chat bubbles (can be moved to CSS or components later)
    const bubbleBaseStyle = "p-3 rounded-lg max-w-[80%]";
    const userBubbleStyle = "bg-blue-500 text-white self-end";
    const aiBubbleStyle = "bg-gray-300 text-gray-800 self-start"; // AI responses use ResultDisplay now

    return (
        <div className="container mx-auto p-4 flex flex-col h-[calc(100vh-theme(spacing.16))]" /* Adjust height based on header */>
            <h1 className="text-2xl font-bold mb-4">RAG Q&A</h1>
            <p className="mb-4 text-gray-600">
                Knowledge Base の情報をもとに、質問に回答します。
            </p>

            {/* Chat History Area */}
            <div className="flex-grow overflow-y-auto mb-4 p-4 bg-gray-100 rounded-lg space-y-4">
                {chatHistory.map((msg, index) => (
                    <div key={index} className={`flex ${msg.sender === 'user' ? 'justify-end' : 'justify-start'}`}>
                        {msg.sender === 'user' && (
                            <div className={cn(bubbleBaseStyle, userBubbleStyle)}>
                                {msg.content}
                            </div>
                        )}
                         {msg.sender === 'ai' && (
                             <ResultDisplay
                                 content={msg.content}
                                 sources={mapSourcesToDocumentSnippets(msg.sources)} // Convert sources before passing
                                 className="max-w-[80%]"
                             />
                         )}
                    </div>
                ))}
                 {isPending && ( // Show spinner while AI is thinking
                     <div className="flex justify-start">
                         <div className={cn(bubbleBaseStyle, aiBubbleStyle, "flex items-center")}>
                              <Spinner size="sm" className="mr-2" />
                              <span>回答生成中...</span>
                         </div>
                     </div>
                 )}
                <div ref={chatEndRef} /> {/* Element to scroll to */}
            </div>

            {/* Input Area */}
            {/* Wrap TextAreaInput and Button for better layout control if needed */}
            <div className="flex items-start gap-2">
                <TextAreaInput
                    value={inputText}
                    onChange={setInputText}
                    onSubmit={handleSendMessage} // Use onSubmit for Enter key sending
                    placeholder="質問を入力してください..."
                    rows={1} // Start with 1 row, auto-grows potentially
                    isLoading={isPending}
                    buttonText="送信" // Use custom button below
                    className="flex-grow"
                />
                 <Button
                     onClick={handleSendMessage}
                     disabled={isPending || !inputText.trim()}
                     isLoading={isPending}
                     className="self-end" // Align button with bottom of text area if it grows
                 >
                    送信
                 </Button>
            </div>
        </div>
    );
};

export default QAPage; 