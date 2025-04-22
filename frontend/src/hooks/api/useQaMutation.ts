import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { QaRequest, QaResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
const qa = async (data: QaRequest): Promise<QaResponse> => {
  return await apiClient.post<QaResponse>('/qa', data);
};

// カスタムフック
export const useQaMutation = () => {
  return useMutation<QaResponse, ApiError, QaRequest>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: qa,
    // 必要に応じて onSuccess, onError などのオプションを設定
    // 例:
    // onSuccess: (data, variables) => {
    //   console.log('QA successful:', data);
    //   // QA成功時の処理 (例: チャット履歴の更新)
    // },
    // onError: (error, variables) => {
    //   console.error('QA failed:', error.message, error.info);
    //   // QA失敗時の処理 (例: エラーメッセージ表示)
    // },
  });
}; 