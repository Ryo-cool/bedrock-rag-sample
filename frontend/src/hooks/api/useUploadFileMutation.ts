import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { UploadFileResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
// ファイルアップロードなので、引数は FormData を想定
const uploadFile = async (formData: FormData): Promise<UploadFileResponse> => {
  return await apiClient.postFormData<UploadFileResponse>('/upload', formData);
};

// カスタムフック
export const useUploadFileMutation = () => {
  return useMutation<UploadFileResponse, ApiError, FormData>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: uploadFile,
    // 必要に応じて onSuccess, onError などのオプションを設定
    // 例:
    // onSuccess: (data) => {
    //   console.log('File uploaded successfully:', data);
    //   // アップロード成功時の処理（例: 通知表示、ファイルリスト更新トリガー）
    // },
    // onError: (error) => {
    //   console.error('File upload failed:', error.message, error.info);
    //   // アップロード失敗時の処理（例: エラー通知）
    // },
  });
}; 