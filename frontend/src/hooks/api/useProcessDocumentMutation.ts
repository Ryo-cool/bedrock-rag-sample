import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { ProcessDocumentRequest, ProcessDocumentResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
const processDocument = async (data: ProcessDocumentRequest): Promise<ProcessDocumentResponse> => {
  return await apiClient.post<ProcessDocumentResponse>('/document/process', data);
};

// カスタムフック
export const useProcessDocumentMutation = () => {
  return useMutation<ProcessDocumentResponse, ApiError, ProcessDocumentRequest>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: processDocument,
    // 必要に応じて onSuccess, onError などのオプションを設定
    // 例:
    // onSuccess: (data, variables) => {
    //   console.log(`Document ${variables.file_id} processing started/completed:`, data);
    //   // ドキュメント処理開始/完了時の処理 (例: ステータス更新、通知)
    // },
    // onError: (error, variables) => {
    //   console.error(`Document ${variables.file_id} processing failed:`, error.message, error.info);
    //   // ドキュメント処理失敗時の処理
    // },
  });
}; 