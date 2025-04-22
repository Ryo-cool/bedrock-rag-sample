import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { SummarizeTextRequest, SummarizeTextResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
const summarizeText = async (data: SummarizeTextRequest): Promise<SummarizeTextResponse> => {
  return await apiClient.post<SummarizeTextResponse>('/summarize/text', data);
};

// カスタムフック
export const useSummarizeTextMutation = () => {
  return useMutation<SummarizeTextResponse, ApiError, SummarizeTextRequest>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: summarizeText,
    // onSuccess, onError, onSettled などのオプションはここで設定可能
    // 例:
    // onSuccess: (data) => {
    //   console.log('Summarization successful:', data);
    // },
    // onError: (error) => {
    //   console.error('Summarization failed:', error.message, error.info);
    //   // ここでエラー通知を表示するなどの処理を追加できる
    // },
  });
}; 