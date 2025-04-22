'use client';

import React, { useState } from 'react';
import { Button } from '@/components/atoms/Button';
// import { Spinner } from '@/components/atoms/Spinner'; // Removed unused import
import { FileUploadInput } from '@/components/molecules/FileUploadInput';
import { ResultDisplay } from '@/components/organisms/ResultDisplay';
import { useUploadFileMutation } from '@/hooks/api/useUploadFileMutation';
// Import other necessary hooks later (e.g., useSummarizeFileMutation)

const UploadPage: React.FC = () => {
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [uploadStatus, setUploadStatus] = useState<string>(''); // To display upload feedback
    const [fileId, setFileId] = useState<string | null>(null); // Store the ID of the uploaded file
    const [summary, setSummary] = useState<string | null>(null); // Store the summary result

    // const { mutate: uploadFile, data: uploadData, error: uploadError, isPending: isUploading } = useUploadFileMutation(); // uploadData removed
    const { mutate: uploadFile, error: uploadError, isPending: isUploading } = useUploadFileMutation();
    // const { mutate: summarizeFile, data: summaryData, error: summaryError, isPending: isSummarizing } = useSummarizeFileMutation(); // Placeholder

    const handleFileSelect = (file: File | null) => {
        setSelectedFile(file);
        setUploadStatus(''); // Reset status on new file select
        setFileId(null);
        setSummary(null);
    };

    const handleUpload = () => {
        if (!selectedFile || isUploading) return;

        setUploadStatus('アップロード中...');
        const formData = new FormData();
        formData.append('file', selectedFile); // 'file' should match the backend expectation

        uploadFile(formData, {
            onSuccess: (data) => {
                setUploadStatus(`アップロード成功！ File ID: ${data.file_id}`);
                setFileId(data.file_id);
                // Optionally trigger summarization or processing immediately
                // handleSummarize(data.file_id);
            },
            onError: (error) => {
                setUploadStatus(`アップロード失敗: ${error.message}`);
            },
        });
    };

    // Placeholder for summarization trigger
    const handleSummarize = (id: string) => {
         if (!id /* || isSummarizing */) return;
         console.log("Triggering summarization for file ID:", id);
        // summarizeFile({ file_id: id }, {
        //     onSuccess: (data) => setSummary(data.summary),
        //     onError: (error) => setSummary(`要約エラー: ${error.message}`),
        // });
    };

     // Determine overall loading state if multiple mutations are involved
     const isLoading = isUploading /* || isSummarizing */;

     // Define accepted file types as a string
     const acceptedTypesString = "application/pdf,image/png,image/jpeg,image/gif,image/bmp,image/webp";

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold mb-4">PDF/画像 アップロード</h1>
            <p className="mb-4 text-gray-600">
                PDF または画像ファイルをアップロードして、内容の要約や処理を行います。
            </p>

            <div className="grid grid-cols-1 gap-6">
                {/* File Upload Area */}
                <div className="space-y-4">
                    <FileUploadInput
                        onFileSelect={handleFileSelect}
                        // acceptedFileTypes={{ 'application/pdf': ['.pdf'], 'image/*': ['.png', '.jpg', '.jpeg', '.gif', '.bmp', '.webp'] }} // Incorrect format
                        acceptedFileTypes={acceptedTypesString} // Corrected: Pass string
                        isLoading={isLoading}
                    />
                     {selectedFile && (
                         <div className="text-sm text-gray-600">
                             選択中のファイル: {selectedFile.name} ({(selectedFile.size / 1024).toFixed(2)} KB)
                         </div>
                     )}
                    <Button
                        onClick={handleUpload}
                        disabled={!selectedFile || isLoading}
                        isLoading={isUploading}
                    >
                        {isUploading ? 'アップロード中...' : 'アップロード実行'}
                    </Button>
                    {/* Add Summarize/Process buttons here, enabled when fileId exists */}
                    {fileId && (
                        <Button
                            onClick={() => handleSummarize(fileId)}
                            // disabled={!fileId || isSummarizing || !!summary}
                            // isLoading={isSummarizing}
                            disabled={true} // Temporarily disabled
                            // variant="outline" // Removed invalid variant
                        >
                            {/* {isSummarizing ? '要約中...' : 'ファイルから要約'} */}
                            ファイルから要約 (未実装)
                        </Button>
                    )}
                </div>

                {/* Result Area */}
                {(uploadStatus || summary) && (
                    <ResultDisplay
                        title="処理結果"
                        content={summary || uploadStatus} // Show summary if available, otherwise upload status
                        isLoading={isLoading}
                        error={uploadError /* || summaryError */} // Combine errors if needed
                    />
                )}
                 {/* Display summary specifically if needed */}
                 {/* {summary && !isSummarizing && !summaryError && (
                     <ResultDisplay title="要約結果" content={summary} />
                 )} */}
            </div>
        </div>
    );
};

export default UploadPage; 