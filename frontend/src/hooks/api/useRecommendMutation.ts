import { useMutation } from '@tanstack/react-query';
import { apiClient, ApiError } from '@/lib/apiClient';
import type { RecommendRequest, RecommendResponse } from '@/types/api';

// Mutation 関数: API コールを行う非同期関数
const recommend = async (data: RecommendRequest): Promise<RecommendResponse> => {
  return await apiClient.post<RecommendResponse>('/recommend', data);
};

// カスタムフック
export const useRecommendMutation = () => {
  return useMutation<RecommendResponse, ApiError, RecommendRequest>({ // <レスポンス型, エラー型, リクエスト型>
    mutationFn: recommend,
    // 必要に応じて onSuccess, onError などのオプションを設定
    // 例:
    // onSuccess: (data, variables) => {
    //   console.log('Recommendation successful:', data);
    //   // 推薦取得成功時の処理
    // },
    // onError: (error, variables) => {
    //   console.error('Recommendation failed:', error.message, error.info);
    //   // 推薦取得失敗時の処理
    // },
  });
}; 