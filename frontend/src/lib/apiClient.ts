// src/lib/apiClient.ts

// カスタムエラークラス
export class ApiError extends Error {
  status: number;
  info?: unknown; // エラーレスポンスの詳細など (any -> unknown)

  constructor(message: string, status: number, info?: unknown) { // any -> unknown
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.info = info;
  }
}

// APIのベースURLを環境変数から取得
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

if (!API_BASE_URL) {
  console.warn(
    '警告: NEXT_PUBLIC_API_BASE_URL が設定されていません。API 通信は失敗します。', 
    '.env.local ファイル等で設定してください。'
  );
}

// 汎用的なリクエスト関数
async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const defaultHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
    // 必要に応じて認証トークンなどを追加
  };

  const config: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
  };

  try {
    const response = await fetch(url, config);

    if (!response.ok) {
      let errorInfo: unknown;
      try {
        errorInfo = await response.json(); // エラーレスポンスボディを試みる
      } catch {
        // JSON パース失敗時はテキストとして取得
        try {
          errorInfo = await response.text();
        } catch {
          errorInfo = 'Failed to get error details'; // テキスト取得も失敗した場合
        }
      }
      throw new ApiError(
        `API Error: ${response.status} ${response.statusText}`, 
        response.status, 
        errorInfo
      );
    }

    // レスポンスボディがない場合 (例: 204 No Content)
    if (response.status === 204) {
      return undefined as T;
    }

    return await response.json() as T;
  } catch (error) {
    if (error instanceof ApiError) {
      // ApiError はそのままスロー
      throw error;
    } else if (error instanceof Error) {
      // その他のネットワークエラーなど
      console.error('Network/Fetch Error:', error);
      throw new ApiError('Network error occurred', 0, { originalError: error.message }); // status 0 などで区別
    } else {
      // 予期せぬエラー
      console.error('Unexpected Error:', error);
      throw new ApiError('An unexpected error occurred', -1); // status -1 などで区別
    }
  }
}

// 各 HTTP メソッド用のヘルパー関数
export const apiClient = {
  get: <T>(endpoint: string, options?: RequestInit) => 
    request<T>(endpoint, { ...options, method: 'GET' }),

  post: <T>(endpoint: string, body: Record<string, unknown>, options?: RequestInit) => 
    request<T>(endpoint, { ...options, method: 'POST', body: JSON.stringify(body) }),

  put: <T>(endpoint: string, body: Record<string, unknown>, options?: RequestInit) => 
    request<T>(endpoint, { ...options, method: 'PUT', body: JSON.stringify(body) }),

  delete: <T>(endpoint: string, options?: RequestInit) => 
    request<T>(endpoint, { ...options, method: 'DELETE' }),

  // ファイルアップロードなど、特殊なケースは別途定義することも可能
  postFormData: <T>(endpoint: string, formData: FormData, options?: RequestInit) => {
    // Content-Type ヘッダーを除外した新しいヘッダーオブジェクトを作成
    const headers: Record<string, string> = {};
    if (options?.headers) {
      for (const [key, value] of Object.entries(options.headers)) {
        if (key.toLowerCase() !== 'content-type') {
          headers[key] = value as string; // Assuming header values are strings
        }
      }
    }
    return request<T>(endpoint, {
      ...options,
      method: 'POST',
      body: formData,
      headers: headers, // Content-Type が除外されたヘッダーを使用
    });
  },
}; 