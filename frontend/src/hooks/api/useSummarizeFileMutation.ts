import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { SummarizeFileRequest, SummarizeFileResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
const summarizeFile = async (data: SummarizeFileRequest): Promise<SummarizeFileResponse> => {
  return await apiClient.post<SummarizeFileResponse>('/summarize/file', data);
};

// カスタムフック
export const useSummarizeFileMutation = () => {
  return useMutation<SummarizeFileResponse, ApiError, SummarizeFileRequest>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: summarizeFile,
    // 必要に応じて onSuccess, onError などのオプションを設定
    // 例:
    // onSuccess: (data, variables) => {
    //   console.log(`File ${variables.file_id} summarized successfully:`, data);
    //   // ファイル要約成功時の処理
    // },
    // onError: (error, variables) => {
    //   console.error(`File ${variables.file_id} summarization failed:`, error.message, error.info);
    //   // ファイル要約失敗時の処理
    // },
  });
}; 